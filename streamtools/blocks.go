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

type STFunc func(inChan chan simplejson.Json, outChan chan simplejson.Json)

type SyncHandler struct {
	msgChan chan simplejson.Json
}

func (self *SyncHandler) HandleMessage(m *nsq.Message) error {
	blob, err := simplejson.NewJson(m.Body)
	if err != nil {
		log.Fatal(err.Error())
	}
	self.msgChan <- *blob
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

func nsqReader(inTopic string, channel string, outChan chan simplejson.Json) {
	r, err := nsq.NewReader(inTopic, channel)
	if err != nil {
		log.Println(inTopic)
		log.Println(channel)
		log.Fatal(err.Error())
	}
	sh := SyncHandler{
		msgChan: outChan,
	}
	r.AddHandler(&sh)
	_ = r.ConnectToLookupd(lookupdHTTPAddrs)
	<-r.ExitChan
}

func TransferBlock(inTopic string, outTopic string, channel string, f STFunc){
	ex := make(chan bool)
	inChan := make(chan simplejson.Json)
	outChan := make(chan simplejson.Json)
	go nsqReader(inTopic, channel, inChan)
	go f(inChan, outChan)
	go nsqWriter(outTopic, outChan)
	<-ex
}