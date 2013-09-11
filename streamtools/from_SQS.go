package streamtools

import (
	"encoding/xml"
	"github.com/bitly/go-simplejson"
	"github.com/bmizerany/aws4"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"time"
)

var (
	AWSAccessKeyId      string = os.Getenv("AWS_ACCESS_KEY_ID")
	AWSAccessSecret     string = os.Getenv("AWS_SECRET_ACCESS_KEY")
	AWSSQSAPIVersion    string = "2012-11-05"
	AWSSignatureVersion string = "4"
)

type Message struct {
	// this is a list in case I'm ever brave enough to up the "MaxNumberOfMessages" away from 1
	Body          []string `xml:"ReceiveMessageResult>Message>Body"`
	ReceiptHandle []string `xml:"ReceiveMessageResult>Message>ReceiptHandle"`
}

func PollSQS(SQSEndpoint string) Message {
	query := make(url.Values)
	query.Add("Action", "ReceiveMessage")
	query.Add("AttributeName", "All")
	query.Add("Version", AWSSQSAPIVersion)
	query.Add("SignatureVersion", AWSSignatureVersion)
	query.Add("WaitTimeSeconds", "10")

	keys := &aws4.Keys{
		AccessKey: AWSAccessKeyId,
		SecretKey: AWSAccessSecret,
	}

	c := aws4.Client{Keys: keys}

	log.Println("[FROMSQS] querying", SQSEndpoint+query.Encode())

	resp, err := c.Get(SQSEndpoint + query.Encode())
	if err != nil {
		log.Fatal(err.Error())
	}

	var v Message

	body, err := ioutil.ReadAll(resp.Body)
	err = xml.Unmarshal(body, &v)
	if err != nil {
		log.Println(err.Error())
	}
	return v

}

func deleteMessage(SQSEndpoint string, ReceiptHandle string) {
	query := make(url.Values)
	query.Add("Action", "DeleteMessage")
	query.Add("ReceiptHandle", ReceiptHandle)
	query.Add("Version", AWSSQSAPIVersion)
	query.Add("SignatureVersion", AWSSignatureVersion)

	keys := &aws4.Keys{
		AccessKey: AWSAccessKeyId,
		SecretKey: AWSAccessSecret,
	}

	c := aws4.Client{Keys: keys}

	log.Println("querying", SQSEndpoint+query.Encode())

	_, err := c.Get(SQSEndpoint + query.Encode())
	if err != nil {
		log.Fatal(err.Error())
	}
}

func FromSQS(outChan chan *simplejson.Json, ruleChan chan *simplejson.Json) {

	log.Println("[FROMSQS] AccessKey:", AWSAccessKeyId)
	log.Println("[FROMSQS] AccessSecret:", AWSAccessSecret)

	rules := <-ruleChan

	SQSEndpoint, err := rules.Get("SQSEndpoint").String()
	if err != nil {
		log.Fatal(err.Error())
	}
	log.Println("[FROMSQS] Listening to", SQSEndpoint)

	timer := time.NewTimer(1)

	for {
		select {
		case <-ruleChan:
		case <-timer.C:
			m := PollSQS(SQSEndpoint)
			if len(m.Body) > 0 {
				for i, body := range m.Body {
					out, err := simplejson.NewJson([]byte(body))
					if err != nil {
						log.Fatal(err.Error())
					}
					outChan <- out
					deleteMessage(SQSEndpoint, m.ReceiptHandle[i])
				}
				timer.Reset(time.Duration(10) * time.Millisecond)
			} else {
				log.Println("[FROMSQS] waiting 10 seconds")
				timer.Reset(time.Duration(10) * time.Second)
			}

		}

	}

}
