package library

import (
	"encoding/json"
	"errors"
	"github.com/bitly/go-nsq"
	"github.com/nytlabs/streamtools/st/blocks" // blocks
	"log"
)

// specify those channels we're going to use to communicate with streamtools
type ToNSQ struct {
	blocks.Block
	queryrule    chan chan interface{}
	inrule       chan interface{}
	in           chan interface{}
	out          chan interface{}
	quit         chan interface{}
	nsqdTCPAddrs string
	topic        string
}

// a bit of boilerplate for streamtools
func NewToNSQ() blocks.BlockInterface {
	return &ToNSQ{}
}

func (b *ToNSQ) Setup() {
	b.Kind = "ToNSQ"
	b.in = b.InRoute("in")
	b.inrule = b.InRoute("rule")
	b.queryrule = b.QueryRoute("rule")
	b.quit = b.InRoute("quit")
	b.out = b.Broadcast()
}

// connects to an NSQ topic and emits each message into streamtools.
func (b *ToNSQ) Run() {
	var writer *nsq.Writer

	for {
		select {
		case ruleI := <-b.inrule:
			// convert message to a map of string interfaces
			rule := ruleI.(map[string]interface{})

			topicI, ok := rule["Topic"]
			if !ok {
				b.Error(errors.New("Topic was not in rule"))
				continue
			}
			topic, ok := topicI.(string)
			if !ok {
				b.Error(errors.New("Topic was not a string"))
				continue
			}

			nsqdTCPAddrsI, ok := rule["NsqdTCPAddrs"]
			if !ok {
				b.Error(errors.New("NsqdTCPAddrs was not in rule"))
				continue
			}
			nsqdTCPAddrs, ok := nsqdTCPAddrsI.(string)
			if !ok {
				b.Error(errors.New("NsqdTCPAddrs was not a string"))
				continue
			}

			writer = nsq.NewWriter(nsqdTCPAddrs)

			b.topic = topic
			b.nsqdTCPAddrs = nsqdTCPAddrs

		case msg := <-b.in:
			log.Println("received message on inroute for topic: ", b.topic)
			log.Println(msg)
			msgStr, err := json.Marshal(msg)
			log.Println("msgStr:", string(msgStr))
			if err != nil {
				b.Error(err)
			}
			frameType, data, err := writer.Publish(b.topic, []byte(msgStr))
			log.Println("frametype:", frameType)
			log.Println("data:", data)
			if err != nil {
				log.Println("error:", err.Error())
				b.Error(err)
			}
			log.Println("done with toNSQ/in")

		case <-b.quit:
			return
		case c := <-b.queryrule:
			c <- map[string]interface{}{
				"Topic":        b.topic,
				"NsqdTCPAddrs": b.nsqdTCPAddrs,
			}
		}
	}
}
