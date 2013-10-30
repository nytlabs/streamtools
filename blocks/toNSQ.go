package blocks

import (
	"encoding/json"
	"github.com/bitly/go-nsq"
	"log"
)

func ToNSQ(b *Block) {

	type toNSQRule struct {
		NsqdHTTPAddrs string
		Topic         string
	}

	var rule *toNSQRule
	var w *nsq.Writer

	// TODO check the endpoint for happiness
	for {
		select {
		case msg := <-b.AddChan:
			updateOutChans(msg, b)
		case <-b.QuitChan:
			quit(b)
			return
		case msg := <-b.InChan:
			if rule == nil {
				break
			}
			blob, err := json.Marshal(msg)
			if err != nil {
				log.Println("failed to marshal JSON")
			}
			frameType, data, err := w.Publish(rule.Topic, blob)
			if err != nil {
				log.Fatalf("frametype %d data %s error %s", frameType, string(data), err.Error())
			}
		case msg := <-b.Routes["set_rule"]:
			if rule == nil {
				rule = &toNSQRule{}
			}
			unmarshal(msg, rule)
			w = nsq.NewWriter(rule.NsqdHTTPAddrs)
		case msg := <-b.Routes["get_rule"]:
			if rule == nil {
				marshal(msg, &toNSQRule{})
			} else {
				marshal(msg, rule)
			}
		}
	}
}
