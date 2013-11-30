package blocks

import (
	"encoding/json"
	"github.com/bitly/go-nsq"
	"log"
)

type readWriteHandler struct {
	toOut chan BMsg
}

func (self readWriteHandler) HandleMessage(message *nsq.Message) error {
	var msg BMsg
	err := json.Unmarshal(message.Body, &msg)
	if err != nil {
		log.Println(err.Error())
	}
	self.toOut <- msg
	return nil
}

func FromNSQ(b *Block) {

	type fromNSQRule struct {
		ReadTopic   string
		LookupdAddr string
		MaxInFlight int
		ReadChannel string
	}

	var rule *fromNSQRule
	var reader *nsq.Reader
	toOut := make(chan BMsg)

	for {
		select {
		case msg := <-toOut:
			broadcast(b.OutChans, msg)
		case msg := <-b.Routes["set_rule"]:
			if rule == nil {
				rule = &fromNSQRule{}
			}
			if reader != nil {
				reader.Stop()
			}

			unmarshal(msg, rule)

			reader, err := nsq.NewReader(rule.ReadTopic, rule.ReadChannel)
			if err != nil {
				log.Fatal(err.Error())
			}
			reader.SetMaxInFlight(rule.MaxInFlight)

			h := readWriteHandler{toOut}
			reader.AddHandler(h)
			err = reader.ConnectToLookupd(rule.LookupdAddr)
			if err != nil {
				log.Fatal(err.Error())
			}
		case msg := <-b.Routes["get_rule"]:
			if rule == nil {
				marshal(msg, &fromNSQRule{})
			} else {
				marshal(msg, rule)
			}
		case msg := <-b.AddChan:
			updateOutChans(msg, b)
		case <-b.QuitChan:
			if reader != nil {
				reader.Stop()
			}
			quit(b)
			return
		}
	}
}
