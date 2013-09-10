package streamtools

import (
	"github.com/bitly/go-nsq"
	"github.com/bitly/go-simplejson"
	"log"
)

var (
	lookupdHTTPAddrs = "127.0.0.1:4161"
	nsqdHTTPAddrs    = "127.0.0.1:4150"
	nsqdTCPAddrs     = "127.0.0.1:4150"
)

type SyncHandler struct {
	msgChan chan *simplejson.Json
}

func (self *SyncHandler) HandleMessage(m *nsq.Message) error {
	blob, err := simplejson.NewJson(m.Body)
	if err != nil {
		log.Fatal(err.Error())
	}
	self.msgChan <- blob
	return nil
}

func nsqReader(topic string, channel string, writeChan chan *simplejson.Json) {
	r, err := nsq.NewReader(topic, channel)
	if err != nil {
		log.Fatal(err.Error())
	}
	sh := SyncHandler{
		msgChan: writeChan,
	}
	r.AddHandler(&sh)
	_ = r.ConnectToLookupd(lookupdHTTPAddrs)
	<-r.ExitChan
}

func nsqWriter(topic string, readChan chan *simplejson.Json) {

	w := nsq.NewWriter(nsqdHTTPAddrs)
	for {
		select {
		case msg := <-readChan:
			outMsg, _ := msg.Encode()
			frameType, data, err := w.Publish(topic, outMsg)
			if err != nil {
				log.Fatalf("frametype %d data %s error %s", frameType, string(data), err.Error())
			}
		}
	}
}

func deMuxWriter(readChan chan *simplejson.Json) {
	w := nsq.NewWriter(nsqdHTTPAddrs)
	for {
		select {
		case msg := <-readChan:
			log.Println(msg)
			topic, err := msg.Get("_StreamtoolsTopic").String()
			if err != nil {
				log.Fatal(err.Error())
			}
			origMsg := msg.Get("_StreamtoolsData")
			log.Println("origMsg:", origMsg)
			outMsg, err := origMsg.Encode()
			if err != nil {
				log.Fatal(err.Error())
			}
			log.Println("outMsg:", string(outMsg))
			frameType, data, err := w.Publish(topic, outMsg)
			if err != nil {
				log.Fatalf("frametype %d data %s error %s", frameType, string(data), err.Error())
			}
		}
	}
}
