package library

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"

	"github.com/mikedewar/aws4"
	"github.com/nytlabs/streamtools/st/blocks" // blocks
	"github.com/nytlabs/streamtools/st/util"
	//"reflect"
	//"strings"
	"sync"
	"time"
)

type sqsMessage struct {
	Body          []string `xml:"ReceiveMessageResult>Message>Body"`
	ReceiptHandle []string `xml:"ReceiveMessageResult>Message>ReceiptHandle"`
}

func dialTimeout(network, addr string) (net.Conn, error) {
	return net.DialTimeout(network, addr, time.Duration(2*time.Second))
}

func (b *FromSQS) listener() {
	log.Println("Starting new SQS listener")
	b.lock.Lock()
	lAuth := map[string]string{}
	var err error
	for k, _ := range b.auth {
		lAuth[k], err = util.ParseString(b.auth, k)
		if err != nil {
			b.Error(err)
			break
		}
	}
	b.listening = true
	b.lock.Unlock()

	transport := http.Transport{
		Dial: dialTimeout,
	}

	httpclient := &http.Client{
		Transport: &transport,
	}

	keys := &aws4.Keys{
		AccessKey: lAuth["AccessKey"],
		SecretKey: lAuth["AccessSecret"],
	}

	sqsclient := &aws4.Client{Keys: keys, Client: httpclient}

	parsedUrl, err := url.Parse(lAuth["SQSEndpoint"])
	if err != nil {
		b.Error(err)
		return
	}

	query := url.Values{}
	query.Set("Action", "ReceiveMessage")
	query.Set("AttributeName", "All")
	query.Set("Version", lAuth["APIVersion"])
	query.Set("SignatureVersion", lAuth["SignatureVersion"])
	query.Set("WaitTimeSeconds", lAuth["WaitTimeSeconds"])
	query.Set("MaxNumberOfMessages", lAuth["MaxNumberOfMessages"])

	parsedUrl.RawQuery = query.Encode()

	queryurl := parsedUrl.String()

	log.Println("Starting SQS read loop")

	for {
		select {
		case <-b.stop:
			log.Println("Exiting SQS read loop")
			return
		default:
			var m sqsMessage

			resp, err := sqsclient.Get(queryurl)

			if err != nil {
				b.Error("could not connect to SQS endpoint. waiting 1s")
				b.Error(err)
				time.Sleep(1 * time.Second)
				continue
			}

			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				b.Error("could not read Body")
				b.Error(err)
				continue
			}

			resp.Body.Close()

			err = xml.Unmarshal(body, &m)
			if err != nil {
				b.Error("could not unmarshal XML")
				b.Error(err)
				continue
			}

			if len(m.Body) == 0 {
				// no messages on queue
				b.Error("no messages on queue. waiting 1s")
				time.Sleep(1 * time.Second)
				continue
			}

			for _, body := range m.Body {
				select {
				case b.fromListener <- []byte(body):
				default:
					log.Println("discarding messages")
					log.Println(len(b.fromListener))
					continue
				}
			}

			parsedUrl, err := url.Parse(lAuth["SQSEndpoint"])
			if err != nil {
				b.Error(err)
				continue
			}

			delquery := url.Values{}
			delquery.Set("Action", "DeleteMessageBatch")
			delquery.Set("Version", lAuth["APIVersion"])
			delquery.Set("SignatureVersion", lAuth["SignatureVersion"])

			for i, r := range m.ReceiptHandle {
				id := fmt.Sprintf("DeleteMessageBatchRequestEntry.%d.Id", (i + 1))
				receipt := fmt.Sprintf("DeleteMessageBatchRequestEntry.%d.ReceiptHandle", (i + 1))
				delquery.Add(id, fmt.Sprintf("msg%d", (i+1)))
				delquery.Add(receipt, r)
			}
			parsedUrl.RawQuery = delquery.Encode()
			delurl := parsedUrl.String()

			resp, err = sqsclient.Get(delurl)
			if err != nil {
				b.Error("could not delete messages. waiting 1s")
				b.Error(err)
				time.Sleep(1 * time.Second)
				continue
			}

			resp.Body.Close()
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
	b.Kind = "Queues"
	b.Desc = "reads from Amazon's SQS, emitting each line of JSON as a separate message"
	b.inrule = b.InRoute("rule")
	b.queryrule = b.QueryRoute("rule")
	b.quit = b.Quit()
	b.out = b.Broadcast()
	b.fromListener = make(chan []byte, 1000)
	b.stop = make(chan bool)
	b.auth = map[string]interface{}{
		"SQSEndpoint":         "",
		"AccessKey":           "",
		"AccessSecret":        "",
		"APIVersion":          "2012-11-05",
		"SignatureVersion":    "4",
		"WaitTimeSeconds":     "0",
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
