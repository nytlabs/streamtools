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
)

type sqsMessage struct {
	Body          []string `xml:"ReceiveMessageResult>Message>Body"`
	ReceiptHandle []string `xml:"ReceiveMessageResult>Message>ReceiptHandle"`
}

func dialTimeout(network, addr string) (net.Conn, error) {
    return net.DialTimeout(network, addr, time.Duration(2 * time.Second))
}

type Reader struct {
	client           *aws4.Client
	sqsEndpoint      string
	version          string
	signatureVersion string
	waitTime         string
	maxMsgs          string
	QuitChan         chan bool   // stops the reader
	OutChan          chan []byte // output channel for the client
}

func NewReader(sqsEndpoint, accessKey, accessSecret string, outChan chan []byte) *Reader {
	transport := http.Transport{
        Dial: dialTimeout,
    }

	client := &http.Client{
        Transport: &transport,
    }

	// ensure that the sqsEndpoint has a ? at the end
	if !strings.HasSuffix(sqsEndpoint, "?") {
		sqsEndpoint += "?"
	}
	AWSSQSAPIVersion := "2012-11-05"
	AWSSignatureVersion := "4"
	keys := &aws4.Keys{
		AccessKey: accessKey,
		SecretKey: accessSecret,
	}
	c := &aws4.Client{Keys: keys, Client: client}
	// channels
	r := &Reader{
		client:           c,
		sqsEndpoint:      sqsEndpoint,
		version:          AWSSQSAPIVersion,
		signatureVersion: AWSSignatureVersion,
		waitTime:         "0",  // in seconds
		maxMsgs:          "10", // in messages
		QuitChan:         make(chan bool),
		OutChan:          outChan,
	}
	return r
}

func readLoop(r *Reader) {

	query := url.Values{}
	query.Set("Action", "ReceiveMessage")
	query.Set("AttributeName", "All")
	query.Set("Version", r.version)
	query.Set("SignatureVersion", r.signatureVersion)
	query.Set("WaitTimeSeconds", r.waitTime)
	query.Set("MaxNumberOfMessages", r.maxMsgs)
	queryurl := r.sqsEndpoint + query.Encode()
	a := time.NewTicker(50 * time.Millisecond)
	for {

		select {
		case <-a.C:
		case <-r.QuitChan:
			log.Println("quitting SQS reader...")
			return
		}
			var m sqsMessage
			var m1 map[string]interface{}

			resp, err := r.client.Get(queryurl)

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
							case r.OutChan <- []byte(outmsg):
							case <-stop.C:
								return
						}
					}(msg)

				}
			}

			delquery := url.Values{}
			delquery.Set("Action", "DeleteMessageBatch")
			delquery.Set("Version", r.version)
			delquery.Set("SignatureVersion", r.signatureVersion)
			for i, r := range m.ReceiptHandle {
				id := fmt.Sprintf("DeleteMessageBatchRequestEntry.%d.Id", (i + 1))
				receipt := fmt.Sprintf("DeleteMessageBatchRequestEntry.%d.ReceiptHandle", (i + 1))
				delquery.Add(id, fmt.Sprintf("msg%d", (i+1)))
				delquery.Add(receipt, r)
			}
			delurl := r.sqsEndpoint + delquery.Encode()

			resp, err = r.client.Get(delurl)
			if err != nil {
				continue
			}

			resp.Body.Close()
	}
}

// specify those channels we're going to use to communicate with streamtools
type FromSQS struct {
	blocks.Block
	queryrule    chan chan interface{}
	inrule       chan interface{}
	out          chan interface{}
	fromReader   chan []byte
	quit         chan interface{}
	SQSEndpoint  string
	AccessKey    string
	AccessSecret string
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
}

// Run is the block's main loop. Here we listen on the different channels we set up.
func (b *FromSQS) Run() {
	var SQSEndpoint, AccessKey, AccessSecret string
	var err error
	var r *Reader
	fromReader := make(chan []byte, 10000)

	for {
		select {
		case msgI := <-b.inrule:
			SQSEndpoint, err = util.ParseString(msgI, "SQSEndpoint")
			if err != nil {
				b.Error(err)
				break
			}

			AccessKey, err = util.ParseString(msgI, "AccessKey")
			if err != nil {
				b.Error(err)
				break
			}

			AccessSecret, err = util.ParseString(msgI, "AccessSecret")
			if err != nil {
				b.Error(err)
				break
			}
			if r != nil {
				r.QuitChan <- true
			}
			r = NewReader(SQSEndpoint, AccessKey, AccessSecret, fromReader)
			go readLoop(r)
		case <-b.quit:
			if r != nil {
				r.QuitChan <- true
			}
			// quit the block
			return
		case msg := <-fromReader:
			var outMsg interface{}
			err := json.Unmarshal(msg, &outMsg)
			if err != nil {
				b.Error(err)
				continue
			}
			b.out <- outMsg
		case respChan := <-b.queryrule:
			// deal with a query request
			respChan <- map[string]interface{}{
				"SQSEndpoint":  SQSEndpoint,
				"AccessKey":    AccessKey,
				"AccessSecret": AccessSecret,
			}
		}
	}
}
