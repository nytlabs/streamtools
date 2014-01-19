package blocks

import (
	"encoding/json"
	"encoding/xml"
	"github.com/mikedewar/aws4"
	"io/ioutil"
	"log"
	"net/url"
	"strings"
	"time"
)

var (
	AWSSQSAPIVersion    string = "2012-11-05"
	AWSSignatureVersion string = "4"
)

type fromSQSRule struct {
	SQSEndpoint  string
	AccessKey    string
	AccessSecret string
}

type Message struct {
	// this is a list in case I'm ever brave enough to up the "MaxNumberOfMessages" away from 1
	Body          []string `xml:"ReceiveMessageResult>Message>Body"`
	ReceiptHandle []string `xml:"ReceiveMessageResult>Message>ReceiptHandle"`
}

func pollSQS(endpoint string, c aws4.Client, rChan chan Message) {
	log.Println("polling start")
	t := time.Now()
	var v Message
	query := make(url.Values)
	query.Add("Action", "ReceiveMessage")
	query.Add("AttributeName", "All")
	query.Add("Version", AWSSQSAPIVersion)
	query.Add("SignatureVersion", AWSSignatureVersion)
	query.Add("WaitTimeSeconds", "0")
	query.Add("MaxNumberOfMessages", "10")

	t1 := time.Now()
	resp, err := c.Get(endpoint + query.Encode())
	log.Println("polling Get", time.Now().Sub(t1).String())
	if err != nil {
		log.Println(err.Error())
		return
	}
	defer resp.Body.Close()
	t1 = time.Now()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err.Error())
		return
	}
	log.Println("polling ReadAll", time.Now().Sub(t1).String())
	t1 = time.Now()
	err = xml.Unmarshal(body, &v)
	if err != nil {
		log.Println(err.Error())
		return
	}
	log.Println("polling Unmarshal", time.Now().Sub(t1).String())
	t1 = time.Now()
	rChan <- v
	log.Println("polling block", time.Now().Sub(t1).String())
	log.Println("polling", time.Now().Sub(t).String())
}

func deleteMessage(endpoint string, c aws4.Client, ReceiptHandle string) {
	query := make(url.Values)
	query.Add("Action", "DeleteMessage")
	query.Add("ReceiptHandle", ReceiptHandle)
	query.Add("Version", AWSSQSAPIVersion)
	query.Add("SignatureVersion", AWSSignatureVersion)
	resp, err := c.Get(endpoint + query.Encode())
	if err != nil {
		log.Println(err.Error())
		return
	}
	resp.Body.Close()
}

func deleteBatch(endpoint string, c aws4.Client, receipts []string) {
	log.Println("delete start")
	t := time.Now()
	query := make(url.Values)
	query.Add("Action", "DeleteMessage")
	query.Add("Version", AWSSQSAPIVersion)
	query.Add("SignatureVersion", AWSSignatureVersion)

	for i, r := range receipts {
		query.Add("DeleteMessageBatchRequestEntry.n.Id", "msg"+string(i))
		query.Add("DeleteMessageBatchRequestEntry.n.ReceiptHandle", r)
	}

	t1 := time.Now()
	resp, err := c.Get(endpoint + query.Encode())
	if err != nil {
		log.Println(err.Error())
	}

	log.Println("delete call", time.Now().Sub(t1).String())
	resp.Body.Close()
	log.Println("deleting", time.Now().Sub(t).String())
}

func unpack(body string, reciept string, endpoint string, client aws4.Client, outChan chan BMsg) {
	log.Println("unpack start")
	t := time.Now()
	var SQSmsg map[string]interface{}
	var msg map[string]interface{}
	err := json.Unmarshal([]byte(body), &SQSmsg)
	if err != nil {
		log.Println(err.Error())
		return
	}
	msgString, ok := SQSmsg["Message"].(string)
	if !ok {
		log.Println("couldn't convert SQS Message to string")
		return
	}
	unpackingBlock := time.Duration(0)
	for _, m := range strings.Split(msgString, "\n") {
		if len(m) == 0 {
			continue
		}
		err = json.Unmarshal([]byte(m), &msg)
		if err != nil {
			log.Println(err.Error())
			continue
		}
		out := BMsg{
			Msg: msg,
		}
		t1 := time.Now()
		outChan <- out
		unpackingBlock = unpackingBlock + time.Now().Sub(t1)
	}
	log.Println("unpacking block", unpackingBlock)
	log.Println("unpacking", time.Now().Sub(t).String())
}

// SQSStream hooks into an Amazon SQS, and emits every message it sees into
// streamtools
func SQSStream(b *Block) {
	var rule *fromSQSRule
	var c aws4.Client
	timer := time.NewTimer(1)
	responseChan := make(chan Message)
	broadcastChan := make(chan BMsg)

	for {
		select {
		case <-timer.C:
			if rule == nil {
				timer.Reset(time.Duration(1) * time.Second)
				break
			}
			go pollSQS(rule.SQSEndpoint, c, responseChan)
			timer.Reset(time.Duration(100) * time.Millisecond)
		case m := <-responseChan:
			if len(m.Body) == 0 {
				timer.Reset(time.Duration(10) * time.Second)
				break
			}
			receipts := make([]string, len(m.Body))
			for i, body := range m.Body {
				reciept := m.ReceiptHandle[i]
				go unpack(body, reciept, rule.SQSEndpoint, c, broadcastChan)
				receipts[i] = reciept
			}
			go deleteBatch(rule.SQSEndpoint, c, receipts)
		case outMsg := <-broadcastChan:
			broadcast(b.OutChans, outMsg)

		case msg := <-b.Routes["set_rule"]:
			if rule == nil {
				rule = &fromSQSRule{}
			}
			unmarshal(msg, rule)
			keys := &aws4.Keys{
				AccessKey: rule.AccessKey,
				SecretKey: rule.AccessSecret,
			}
			c = aws4.Client{Keys: keys}
		case msg := <-b.Routes["get_rule"]:
			if rule == nil {
				marshal(msg, &fromSQSRule{})
			} else {
				marshal(msg, rule)
			}
		case msg := <-b.AddChan:
			updateOutChans(msg, b)
		case <-b.QuitChan:
			quit(b)
			return
		}
	}
}
