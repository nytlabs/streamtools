package streamtools

import (
	"github.com/bitly/go-simplejson"
	"github.com/bitly/nsq/nsq"
	"log"
)

type readWriteHandler struct {
	outChan chan *simplejson.Json
}

func (self readWriteHandler) HandleMessage(message *nsq.Message) error {
	out, err := simplejson.NewJson(message.Body)
	if err != nil {
		log.Fatal(err.Error())
	}
	self.outChan <- out
	return nil
}

func FromNSQ(outChan chan *simplejson.Json, ruleChan chan *simplejson.Json) {

	rules := <-ruleChan

	readTopic, err := rules.Get("readTopic").String()
	lookupdAddr, err := rules.Get("lookupdAddr").String()

	reader, err := nsq.NewReader(readTopic, "fromNSQ")
	if err != nil {
		log.Fatal(err.Error())
	}

	h := readWriteHandler{outChan}
	reader.AddHandler(h)
	err = reader.ConnectToLookupd(lookupdAddr)
	if err != nil {
		log.Fatal(err.Error())
	}
	for {
		select {
		case <-reader.ExitChan:
			break
		case <-ruleChan:
		}
	}
}
