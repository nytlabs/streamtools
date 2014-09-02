package library

import (
	"encoding/json"
	"log"

	"github.com/goamz/goamz/aws"
	"github.com/goamz/goamz/sqs"

	"strconv"
	"sync"

	"github.com/nytlabs/streamtools/st/blocks" // blocks
	"github.com/nytlabs/streamtools/st/util"
)

// lots of this code stolen brazenly from JP https://github.com/jprobinson
func (b *FromSQS) listener() {
	accessKey, ok := b.auth["AccessKey"].(string)
	if !ok {
		b.Error("could not assert AccessKey to a string")
		return
	}
	accessSecret, ok := b.auth["AccessSecret"].(string)
	if !ok {
		b.Error("could not assert AccessSecret to a string")
		return
	}
	queueName, ok := b.auth["QueueName"].(string)
	if !ok {
		b.Error("could not assert queue name to a string")
		return
	}
	auth := aws.Auth{AccessKey: accessKey, SecretKey: accessSecret}
	sqsClient := sqs.New(auth, aws.USEast)
	queue, err := sqsClient.GetQueue(queueName)

	maxNstring, ok := b.auth["MaxNumberOfMessages"].(string)
	if !ok {
		b.Error("could not assert MaxNumberOfMessages to a string")
		return
	}
	maxN, err := strconv.Atoi(maxNstring)
	if err != nil {
		b.Error(err)
		return
	}

	var resp *sqs.ReceiveMessageResponse
	for {
		select {
		case <-b.stop:
			log.Println("Exiting SQS read loop")
			return
		default:
			resp, err = queue.ReceiveMessage(maxN)
			if err != nil {
				b.Error(err)
			}
			if len(resp.Messages) == 0 {
				break
			}
			for _, m := range resp.Messages {
				select {
				case b.fromListener <- []byte(m.Body):
				default:
					log.Println("discarding messages")
					log.Println(len(b.fromListener))
					continue
				}

				if _, err = queue.DeleteMessage(&m); err != nil {
					b.Error(err)
				}
			}
		}
	}
}

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
	auth         map[string]interface{}
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
	b.auth = map[string]interface{}{
		"QueueName":           "",
		"AccessKey":           "",
		"AccessSecret":        "",
		"MaxNumberOfMessages": "10",
	}
}

func (b *FromSQS) stopListening() {
	log.Println("attempting to stop SQS reader")
	if b.listening {
		b.stop <- true
		b.listening = false
	}
}

// Run is the block's main loop. Here we listen on the different channels we set up.
func (b *FromSQS) Run() {
	var err error

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

			b.stopListening()
			go b.listener()
		case <-b.quit:
			b.stopListening()
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
