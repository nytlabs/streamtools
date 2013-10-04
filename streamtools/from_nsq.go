package streamtools

import (
	"github.com/bitly/go-nsq"
	"github.com/bitly/go-simplejson"
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
	maxInFlight, err := rules.Get("maxInFlight").Int()
	readChannel, err := rules.Get("readChannel").String()

	reader, err := nsq.NewReader(readTopic, readChannel)
	if err != nil {
		log.Fatal(err.Error())
	}
	reader.SetMaxInFlight(maxInFlight)

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
