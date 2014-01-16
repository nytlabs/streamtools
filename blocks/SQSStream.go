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
	var v Message
	query := make(url.Values)
	query.Add("Action", "ReceiveMessage")
	query.Add("AttributeName", "All")
	query.Add("Version", AWSSQSAPIVersion)
	query.Add("SignatureVersion", AWSSignatureVersion)
	query.Add("WaitTimeSeconds", "10")
	query.Add("MaxNumberOfMessages", "10")

	resp, err := c.Get(endpoint + query.Encode())
	defer resp.Body.Close()
	if err != nil {
		log.Println(err.Error())
		return
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err.Error())
		return
	}
	err = xml.Unmarshal(body, &v)
	if err != nil {
		log.Println(err.Error())
		return
	}
	rChan <- v
}

func deleteMessage(endpoint string, c aws4.Client, ReceiptHandle string) {
	query := make(url.Values)
	query.Add("Action", "DeleteMessage")
	query.Add("ReceiptHandle", ReceiptHandle)
	query.Add("Version", AWSSQSAPIVersion)
	query.Add("SignatureVersion", AWSSignatureVersion)
	resp, err := c.Get(endpoint + query.Encode())
	resp.Body.Close()
	if err != nil {
		log.Println(err.Error())
	}
}

// SQSStream hooks into an Amazon SQS, and emits every message it sees into
// streamtools
func SQSStream(b *Block) {
	var rule *fromSQSRule
	var c aws4.Client
	var SQSmsg map[string]interface{}
	var msg map[string]interface{}
	timer := time.NewTimer(1)
	responseChan := make(chan Message)

	for {
		select {
		case <-timer.C:
			if rule == nil {
				timer.Reset(time.Duration(10) * time.Second)
				break
			}
			go pollSQS(rule.SQSEndpoint, c, responseChan)
			timer.Reset(time.Duration(1) * time.Millisecond)
		case m := <-responseChan:
			if len(m.Body) == 0 {
				timer.Reset(time.Duration(10) * time.Second)
				break
			}
			log.Println("body length", len(m.Body))

			for i, body := range m.Body {
				err := json.Unmarshal([]byte(body), &SQSmsg)
				msgString, ok := SQSmsg["Message"].(string)
				if !ok {
					log.Println("couldn't convert SQS Message to string")
					continue
				}
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
					broadcast(b.OutChans, out)
				}
				go deleteMessage(rule.SQSEndpoint, c, m.ReceiptHandle[i])
			}

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
