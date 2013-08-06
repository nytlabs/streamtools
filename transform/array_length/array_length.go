package main

import (
	"flag"
	"github.com/bitly/go-simplejson"
	"github.com/bitly/nsq/nsq"
	"log"
)

var (
	inTopic          = flag.String("in_topic", "", "topic to read from")
	outTopic         = flag.String("out_topic", "", "topic to write to")
	arrayKey         = flag.String("key", "", "key of the array whose length you would like")
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

func GetLength(msgChan chan *nsq.Message, outChan chan int) {
	for {
		select {
		case m := <-msgChan:
			blob, err := simplejson.NewJson(m.Body)
			if err != nil {
				log.Fatal(err.Error())
			}
			arr, err := blob.Get(*arrayKey).Array()
			if err != nil {
				log.Fatal(err.Error())
			}
			l := len(arr)
			outChan <- l
		}
	}
}

func Writer(outChan chan int) {
	w := nsq.NewWriter(0)
	err := w.ConnectToNSQ(nsqdAddr)
	if err != nil {
		log.Fatal(err.Error())
	}
	for {
		select {
		case l := <-outChan:
            // TODO make this simplejson.NewJson([]byte{'{}'})
            msg,_ := simplejson.NewJson([]byte("{}"))
            msg.Set("len_" + *arrayKey, l)
            outMsg, _ := msg.Encode()
			frameType, data, err := w.Publish(*outTopic, outMsg)
			if err != nil {
				log.Fatalf("frametype %d data %s error %s", frameType, string(data), err.Error())
			}
		}
	}

}

func main() {

	flag.Parse()
	channel := "length_" + *arrayKey
	r, err := nsq.NewReader(*inTopic, channel)
	if err != nil {
		log.Println(*inTopic)
		log.Println(channel)
		log.Fatal(err.Error())
	}

	msgChan := make(chan *nsq.Message)
	outChan := make(chan int)

	go GetLength(msgChan, outChan)
	go Writer(outChan)

	sh := SyncHandler{
		msgChan: msgChan,
	}
	r.AddHandler(&sh)
	_ = r.ConnectToLookupd(lookupdHTTPAddrs)

	<-r.ExitChan
}
