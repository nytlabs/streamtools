package blocks

import (
	"github.com/mikedewar/go-sqsReader" //sqsReader
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

// SQSStream hooks into an Amazon SQS, and emits every message it sees into
// streamtools
func SQSStream(b *Block) {
	var rule *fromSQSRule
	var r *sqsReader.Reader
	outChan := make(chan map[string]interface{})

	for {
		select {
		case m := <-outChan:
			out := BMsg{
				Msg: m,
			}
			broadcast(b.OutChans, out)
		case msg := <-b.Routes["set_rule"]:
			if rule == nil {
				rule = &fromSQSRule{}
			}
			unmarshal(msg, rule)
			r = sqsReader.NewReader(rule.SQSEndpoint, rule.AccessKey, rule.AccessSecret, outChan)
			go r.Start()
		case msg := <-b.Routes["get_rule"]:
			if rule == nil {
				marshal(msg, &fromSQSRule{})
			} else {
				marshal(msg, rule)
			}
		case msg := <-b.AddChan:
			updateOutChans(msg, b)
		case <-b.QuitChan:
			r.Stop()
			quit(b)
			return
		}
	}
}
