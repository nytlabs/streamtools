package blocks

import (
	"github.com/bitly/go-nsq"
	"encoding/json"
	"log"
)

type readWriteHandler struct {
	OutChans map[string]chan BMsg
}

func (self readWriteHandler) HandleMessage(message *nsq.Message) error {
	var msg BMsg 
	err := json.Unmarshal(message.Body, &msg)
	if err != nil {
		log.Println(err.Error())
	}	

	broadcast(self.OutChans, msg)
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
		case msg := <-b.AddChan:
			updateOutChans(msg, b)
		case <-b.QuitChan:
			reader.ExitChan <- 1
			quit(b)
			return
		}
	}
}
