package library

import (
	"encoding/json"
	"github.com/nytlabs/streamtools/st/blocks" // blocks
	"github.com/nytlabs/streamtools/st/util"
	"encoding/xml"
	"fmt"
	"github.com/mikedewar/aws4"
	"io/ioutil"
	"net/url"
	"strings"
	"log"
	"time"
	"net"
	"net/http"
    "sync"
)

type sqsMessage struct {
	Body          []string `xml:"ReceiveMessageResult>Message>Body"`
	ReceiptHandle []string `xml:"ReceiveMessageResult>Message>ReceiptHandle"`
}

func dialTimeout(network, addr string) (net.Conn, error) {
    return net.DialTimeout(network, addr, time.Duration(2 * time.Second))
}

func (b *FromSQS) listener() {
	b.lock.Lock()
	lAuth := map[string]string{}
	for k, _ := range b.auth {
		lAuth[k] = b.auth[k]
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
			var m1 map[string]interface{}

			resp, err := sqsclient.Get(queryurl)

			if err != nil {
				continue
			}

			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				continue
			}

			resp.Body.Close()

			err = xml.Unmarshal(body, &m)
			if err != nil {
				continue
			}

			if len(m.Body) == 0 {
				continue
			}

			for _, body := range m.Body {
				err = json.Unmarshal([]byte(body), &m1)
				if err != nil {
					continue
				}
				msgString, ok := m1["Message"].(string)
				if !ok {
					continue
				}
				msgs := strings.Split(msgString, "\n")
				for _, msg := range msgs {
					if len(msg) == 0 {
						continue
					}

					go func(outmsg string){
						stop := time.NewTimer(1 * time.Second)
						select {
							case b.fromListener <- []byte(outmsg):
							case <-stop.C:
								return
						}
					}(msg)

				}
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
				continue
			}

			resp.Body.Close()
    	}
    }
}


// specify those channels we're going to use to communicate with streamtools
type FromSQS struct {
	blocks.Block
	queryrule    chan chan interface{}
	inrule       chan interface{}
	out          chan interface{}
	quit         chan interface{}

    lock         sync.Mutex
    listening    bool
    fromListener chan []byte
    auth         map[string]string
    stop 		 chan bool
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
	b.auth = map[string]string{
		"SQSEndpoint":  "",
		"AccessKey":    "",
		"AccessSecret": "",
		"APIVersion":      "2012-11-05",
		"SignatureVersion": "4",
		"WaitTimeSeconds": "0",
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
