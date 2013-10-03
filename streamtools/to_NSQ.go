package streamtools

import (
	"github.com/bitly/go-nsq"
	"github.com/bitly/go-simplejson"
	"log"
)

func ToNSQ(inChan chan *simplejson.Json, ruleChan chan *simplejson.Json) {
	rule := <-ruleChan
	nsqdHTTPAddrs, err := rule.Get("nsqdHTTPAddrs").String()
	if err != nil {
		log.Fatal(err.Error())
	}
	topic, err := rule.Get("topic").String()
	if err != nil {
		log.Fatal(err.Error())
	}
	w := nsq.NewWriter(nsqdHTTPAddrs)

	for {
		select {
		case <-ruleChan:
		case msg := <-inChan:
			outMsg, err := msg.Encode()
			if err != nil {
				log.Fatal(err.Error())
			}
			frameType, data, err := w.Publish(topic, outMsg)
			if err != nil {
				log.Fatalf("frametype %d data %s error %s", frameType, string(data), err.Error())
			}
		}
	}

}
