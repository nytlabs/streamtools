package blocks

import (
	"github.com/bitly/go-nsq"
	"github.com/bitly/go-simplejson"
	"log"
)

type readWriteHandler struct {
	OutChans map[string]chan *simplejson.Json
}

func (self readWriteHandler) HandleMessage(message *nsq.Message) error {
	out, err := simplejson.NewJson(message.Body)
	if err != nil {
		log.Fatal(err.Error())
	}
	broadcast(self.OutChans, out)
	return nil
}

func FromNSQ(b *Block) {

	type fromNSQRule struct {
		ReadTopic   string
		LookupdAddr string
		MaxInFlight int
		ReadChannel string
	}

	rule := &fromNSQRule{}

	unmarshal(<-b.Routes["set_rule"], &rule)

	reader, err := nsq.NewReader(rule.ReadTopic, rule.ReadChannel)
	if err != nil {
		log.Fatal(err.Error())
	}
	reader.SetMaxInFlight(rule.MaxInFlight)

	h := readWriteHandler{b.OutChans}
	reader.AddHandler(h)
	err = reader.ConnectToLookupd(rule.LookupdAddr)
	if err != nil {
		log.Fatal(err.Error())
	}
	for {
		select {
		case <-reader.ExitChan:
			break
		}
	}
}
