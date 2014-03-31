package library

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"github.com/mikedewar/aws4"
	"github.com/nytlabs/streamtools/st/blocks" // blocks
	"github.com/nytlabs/streamtools/st/util"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
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

	query := url.Values{}
	query.Set("Action", "ReceiveMessage")
	query.Set("AttributeName", "All")
	query.Set("Version", lAuth["APIVersion"])
	query.Set("SignatureVersion", lAuth["SignatureVersion"])
	query.Set("WaitTimeSeconds", lAuth["WaitTimeSeconds"])
	query.Set("MaxNumberOfMessages", lAuth["MaxNumberOfMessages"])
	queryurl := lAuth["SQSEndpoint"] + query.Encode()

	for {
		select {
		case <-b.stop:
			log.Println("Exiting SQS read loop")
			return
		default:
			var m sqsMessage
			//var m1 map[string]interface{}

			resp, err := sqsclient.Get(queryurl)

			if err != nil {
				b.Error("could not connect to SQS endpoint")
				b.Error(err)
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
				time.Sleep(1 * time.Second)
				//log.Println("sleeping for a second")
				continue
			}

			for _, body := range m.Body {
				/*
					err = json.Unmarshal([]byte(body), &m1)
					if err != nil {
						b.Error("could not unmarshal JSON")
						b.Error(err)
						continue
					}
							message, ok := m1[unpackPath]
							if !ok {
								b.Error("could not find", unpackPath, "in JSON")
								continue
							}

							msgString, ok := message.(string)
							if !ok {
								log.Println(message)
								b.Error("could not assert Message to string")
								b.Error(err)
								continue
							}
							msgs := strings.Split(msgString, "\n")
						for _, msg := range msgs {
							if len(msg) == 0 {
								continue
							}
				*/

				go func(outmsg string) {
					stop := time.NewTimer(1 * time.Second)
					select {
					case b.fromListener <- []byte(outmsg):
					case <-stop.C:
						return
					}
				}(body)
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
			delurl := lAuth["SQSEndpoint"] + delquery.Encode()

			resp, err = sqsclient.Get(delurl)
			if err != nil {
				b.Error("could not delete messages")
				b.Error(err)
				continue
			}

			resp.Body.Close()
		}
	}
}

// specify those channels we're going to use to communicate with streamtools
type FromSQS struct {
	blocks.Block
	queryrule chan chan interface{}
	inrule    chan interface{}
	out       chan interface{}
	quit      chan interface{}

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
	b.Kind = "fromSQS"
	b.inrule = b.InRoute("rule")
	b.queryrule = b.QueryRoute("rule")
	b.quit = b.Quit()
	b.out = b.Broadcast()
	b.fromListener = make(chan []byte)
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
		case respChan := <-b.queryrule:
			// deal with a query request
			respChan <- b.auth
		}
	}
}
