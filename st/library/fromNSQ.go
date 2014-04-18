package library

import (
	"encoding/json"
	"github.com/bitly/go-nsq"
	"github.com/nytlabs/streamtools/st/blocks"
	"github.com/nytlabs/streamtools/st/util"
)

// specify those channels we're going to use to communicate with streamtools
type FromNSQ struct {
	blocks.Block
	queryrule chan chan interface{}
	inrule    chan interface{}
	out       chan interface{}
	quit      chan interface{}
}

// a bit of boilerplate for streamtools
func NewFromNSQ() blocks.BlockInterface {
	return &FromNSQ{}
}

func (b *FromNSQ) Setup() {
	b.Kind = "FromNSQ"
	b.Desc = "reads from a topic in NSQ as specified in this block's rule"
	b.inrule = b.InRoute("rule")
	b.queryrule = b.QueryRoute("rule")
	b.quit = b.Quit()
	b.out = b.Broadcast()
}

type readWriteHandler struct {
	toOut   chan interface{}
	toError chan error
}

func (self readWriteHandler) HandleMessage(message *nsq.Message) error {
	var msg interface{}
	err := json.Unmarshal(message.Body, &msg)
	if err != nil {
		msg = map[string]interface{}{
			"data": message.Body,
		}
	}
	self.toOut <- msg
	return nil
}

// connects to an NSQ topic and emits each message into streamtools.
func (b *FromNSQ) Run() {
	var reader *nsq.Reader
	var topic, channel, lookupdAddr string
	var maxInFlight float64
	var err error
	toOut := make(chan interface{})
	toError := make(chan error)

	for {
		select {
		case msg := <-toOut:
			b.out <- msg
		case err := <-toError:
			b.Error(err)
		case ruleI := <-b.inrule:
			// convert message to a map of string interfaces
			// aka keys are strings, values are empty interfaces
			rule := ruleI.(map[string]interface{})

			topic, err = util.ParseString(rule, "ReadTopic")
			if err != nil {
				b.Error(err)
				continue
			}

			lookupdAddr, err = util.ParseString(rule, "LookupdAddr")
			if err != nil {
				b.Error(err)
				continue
			}
			maxInFlight, err = util.ParseFloat(rule, "MaxInFlight")
			if err != nil {
				b.Error(err)
				continue
			}
			channel, err = util.ParseString(rule, "ReadChannel")
			if err != nil {
				b.Error(err)
				continue
			}

			if reader != nil {
				reader.Stop()
			}

			reader, err = nsq.NewReader(topic, channel)
			if err != nil {
				b.Error(err)
				continue
			}
			reader.SetMaxInFlight(int(maxInFlight))

			h := readWriteHandler{toOut, toError}
			reader.AddHandler(h)

			err = reader.ConnectToLookupd(lookupdAddr)
			if err != nil {
				b.Error(err)
				continue
			}

		case <-b.quit:
			if reader != nil {
				reader.Stop()
			}
			return
		case c := <-b.queryrule:
			c <- map[string]interface{}{
				"ReadTopic":   topic,
				"ReadChannel": channel,
				"LookupdAddr": lookupdAddr,
				"MaxInFlight": maxInFlight,
			}
		}
	}
}
