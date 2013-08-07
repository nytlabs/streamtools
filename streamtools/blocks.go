package streamtools

import (
	"github.com/bitly/go-simplejson"
	"github.com/bitly/nsq/nsq"
	"log"
)

var (
	lookupdHTTPAddrs = "127.0.0.1:4161"
	nsqdAddr         = "127.0.0.1:4150"
)

type SyncHandler struct {
	msgChan chan *nsq.Message
}

func (self *SyncHandler) HandleMessage(m *nsq.Message) error {
	self.msgChan <- m
	return nil
}

func nsqWriter(outTopic string, outChan chan simplejson.Json) {
	w := nsq.NewWriter(0)
	err := w.ConnectToNSQ(nsqdAddr)
	if err != nil {
		log.Fatal(err.Error())
	}
	for {
		select {
		case l := <-outChan:
			outMsg, _ := l.Encode()
			frameType, data, err := w.Publish(outTopic, outMsg)
			if err != nil {
				log.Fatalf("frametype %d data %s error %s", frameType, string(data), err.Error())
			}
		}
	}
}

func TransferBlock(inTopic string, outTopic string, channel string, f func(msgChan chan *nsq.Message, outChan chan simplejson.Json)) {
	r, err := nsq.NewReader(inTopic, channel)
	if err != nil {
		log.Println(inTopic)
		log.Println(channel)
		log.Fatal(err.Error())
	}
	msgChan := make(chan *nsq.Message)
	outChan := make(chan simplejson.Json)
	go f(msgChan, outChan)
	go nsqWriter(outTopic, outChan)
	sh := SyncHandler{
		msgChan: msgChan,
	}
	r.AddHandler(&sh)
	_ = r.ConnectToLookupd(lookupdHTTPAddrs)
	<-r.ExitChan
}