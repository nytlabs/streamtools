package library

import (
	"encoding/json"
	"errors"
	"github.com/bitly/go-nsq"
	"github.com/nytlabs/streamtools/st/blocks" // blocks
)

// specify those channels we're going to use to communicate with streamtools
type FromNSQ struct {
	blocks.Block
	queryrule   chan chan interface{}
	inrule      chan interface{}
	out         chan interface{}
	quit        chan interface{}
	topic       string
	channel     string
	lookupdAddr string
	maxInFlight int
}

// a bit of boilerplate for streamtools
func NewFromNSQ() blocks.BlockInterface {
	return &FromNSQ{}
}

func (b *FromNSQ) Setup() {
	b.Kind = "FromNSQ"
	b.inrule = b.InRoute("rule")
	b.queryrule = b.QueryRoute("rule")
	b.quit = b.InRoute("quit")
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
		self.toError <- err
		return err
	}
	self.toOut <- msg
	return nil
}

// connects to an NSQ topic and emits each message into streamtools.
func (b *FromNSQ) Run() {
	var reader *nsq.Reader
	toOut := make(chan interface{})
	toError := make(chan error)

	for {
		select {
		case msg := <-toOut:
			b.out <- &msg
		case err := <-toError:
			b.Error(err)
		case ruleI := <-b.inrule:
			// convert message to a map of string interfaces
			rule := ruleI.(map[string]interface{})

			// TODO: make this pattern a util so we don't have to copy/paste so much code
			topicI, ok := rule["ReadTopic"]
			if !ok {
				b.Error(errors.New("ReadTopic was not in rule"))
				continue
			}
			topic, ok := topicI.(string)
			if !ok {
				b.Error(errors.New("topic was not a string"))
				continue
			}

			lookupdAddrI, ok := rule["LookupdAddr"]
			if !ok {
				b.Error(errors.New("LookupdAddr was not in rule"))
				continue
			}
			lookupdAddr, ok := lookupdAddrI.(string)
			if !ok {
				b.Error(errors.New("LookupdAddr was not a string"))
				continue
			}

			maxInFlightI, ok := rule["MaxInFlight"]
			if !ok {
				b.Error(errors.New("ReadmaxInFlight was not in rule"))
				continue
			}
			maxInFlight, ok := maxInFlightI.(int)
			if !ok {
				b.Error(errors.New("MaxInFlight was not an integer"))
				continue
			}

			channelI, ok := rule["ReadChannel"]
			if !ok {
				b.Error(errors.New("ReadChannel was not in rule"))
				continue
			}
			channel, ok := channelI.(string)
			if !ok {
				b.Error(errors.New("channel was not a string"))
				continue
			}

			reader, err := nsq.NewReader(topic, channel)
			if err != nil {
				b.Error(err)
			}
			reader.SetMaxInFlight(maxInFlight)

			h := readWriteHandler{toOut, toError}
			reader.AddHandler(h)

			err = reader.ConnectToLookupd(lookupdAddr)
			if err != nil {
				b.Error(err)
			}

			b.topic = topic
			b.channel = channel
			b.maxInFlight = maxInFlight
			b.lookupdAddr = lookupdAddr

		case <-b.quit:
			if reader != nil {
				reader.Stop()
			}
			return
		case c := <-b.queryrule:
			c <- map[string]interface{}{
				"ReadTopic":   b.topic,
				"ReadChannel": b.channel,
				"LookupdAddr": b.lookupdAddr,
				"MaxInFlight": b.maxInFlight,
			}
		}
	}
}
