package library

import (
	"encoding/json"
	"errors"
	"log"
	"time"

	"github.com/goamz/goamz/aws"
	"github.com/goamz/goamz/sqs"

	"sync"

	"github.com/nytlabs/streamtools/st/blocks" // blocks
	"github.com/nytlabs/streamtools/st/util"
)


// specify those channels we're going to use to communicate with streamtools
type FromSQS struct {
	blocks.Block
	queryrule chan blocks.MsgChan
	inrule    blocks.MsgChan
	out       blocks.MsgChan
	quit      blocks.MsgChan

	lock         sync.Mutex
	listening    bool
	fromListener chan []byte
	auth         map[string]string
	stop         chan bool
}

// we need to build a simple factory so that streamtools can make new blocks of this kind
func NewFromSQS() blocks.BlockInterface {
	return &FromSQS{}
}

// Setup is called once before running the block. We build up the channels and specify what kind of block this is.
func (b *FromSQS) Setup() {
	b.Kind = "Queue I/O"
	b.Desc = "reads from Amazon's SQS, emitting each line of JSON as a separate message"
	b.inrule = b.InRoute("rule")
	b.queryrule = b.QueryRoute("rule")
	b.quit = b.Quit()
	b.out = b.Broadcast()
	b.fromListener = make(chan []byte, 1000)
	b.stop = make(chan bool)
	b.auth = map[string]string{
		"QueueName":           "",
		"AccessKey":           "",
		"AccessSecret":        "",
		"MaxNumberOfMessages": "10",
	}
}

func (b *FromSQS) runReader(sem chan bool, outChan chan []byte, stopChan chan bool, auth map[string]string) {
	log.Println("starting new reader")
	t := time.Now()
	err := listener(auth["AccessKey"], auth["AccessSecret"], auth["QueueName"], auth["MaxNumberOfMessages"], outChan, stopChan)
	if err != nil {
		b.Error(err)
		if time.Since(t) < 1*time.Second {
			log.Println("reader failed in less than one second")
			b.Error(errors.New("reader died rapidly - check SQS reader parameters"))
			// here we don't free up a seperate reader
		} else {
			log.Println("freeing reader")
			<-sem
		}
	}

}

func stopAllReaders(stopChans []chan bool) {
	for _, stopChan := range stopChans {
		close(stopChan)
	}
}

// lots of this code stolen brazenly from JP https://github.com/jprobinson
func listener(key, secret, queueName, maxN string, outChan chan []byte, stopChan chan bool) error {
	auth := aws.Auth{AccessKey: key, SecretKey: secret}
	sqsClient := sqs.New(auth, aws.USEast)
	log.Println("getting SQS queue")
	queue, err := sqsClient.GetQueue(queueName)
	if err != nil {
		return err
	}

	params := map[string]string{
		"WaitTimeSeconds":     "1",
		"MaxNumberOfMessages": maxN,
	}

	var resp *sqs.ReceiveMessageResponse
	log.Println("starting read loop")
	for {
		select {
		case <-stopChan:
			log.Println("Exiting SQS read loop")
			return nil
		default:
			resp, err = queue.ReceiveMessageWithParameters(params)
			if err != nil {
				return err
			}
			if len(resp.Messages) == 0 {
				break
			}
			for _, m := range resp.Messages {
				select {
				case outChan <- []byte(m.Body):
				default:
					log.Println("discarding messages")
					log.Println(len(outChan))
					continue
				}
				_, err = queue.DeleteMessage(&m)
				if err != nil {
					return err
				}
			}
		}
	}
}

// Run is the block's main loop. Here we listen on the different channels we set up.
func (b *FromSQS) Run() {
	var err error
	numReaders := 10
	stopChans := make([]chan bool, 0)
	semChan := make(chan bool, numReaders)

	for {
		select {
		case msgI := <-b.inrule:
			for k, _ := range b.auth {
				b.auth[k], err = util.ParseString(msgI, k)
				if err != nil {
					b.Error(err)
					break
				}
			}

			stopAllReaders(stopChans)
			stopChans = make([]chan bool, 0)

			go func() {
				for {
					semChan <- true // maintain pressure
					stopChan := make(chan bool, 1)
					stopChans = append(stopChans, stopChan)
					go b.runReader(semChan, b.fromListener, stopChan, b.auth)
				}
			}()

		case <-b.quit:
			stopAllReaders(stopChans)
			return
		case msg := <-b.fromListener:
			var outMsg interface{}
			err := json.Unmarshal(msg, &outMsg)
			if err != nil {
				b.Error(err)
				continue
			}
			b.out <- outMsg
		case MsgChan := <-b.queryrule:
			// deal with a query request
			MsgChan <- b.auth
		}
	}
}
