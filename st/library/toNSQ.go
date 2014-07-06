package library

import (
	"encoding/json"

	"github.com/bitly/go-nsq"
	"github.com/nytlabs/streamtools/st/blocks" // blocks
	"github.com/nytlabs/streamtools/st/util"
)

// specify those channels we're going to use to communicate with streamtools
type ToNSQ struct {
	blocks.Block
	queryrule chan blocks.MsgChan
	inrule    blocks.MsgChan
	in        blocks.MsgChan
	out       blocks.MsgChan
	quit      blocks.MsgChan
}

// a bit of boilerplate for streamtools
func NewToNSQ() blocks.BlockInterface {
	return &ToNSQ{}
}

func (b *ToNSQ) Setup() {
	b.Kind = "Queues"
	b.Desc = "send messages to an NSQ topic"
	b.in = b.InRoute("in")
	b.inrule = b.InRoute("rule")
	b.queryrule = b.QueryRoute("rule")
	b.quit = b.Quit()
}

// connects to an NSQ topic and emits each message into streamtools.
func (b *ToNSQ) Run() {
	var err error
	var nsqdTCPAddrs string
	var topic string
	var writer *nsq.Producer

	conf := nsq.NewConfig()

	for {
		select {
		case ruleI := <-b.inrule:
			topic, err = util.ParseString(ruleI, "Topic")
			if err != nil {
				b.Error(err)
				break
			}

			nsqdTCPAddrs, err = util.ParseString(ruleI, "NsqdTCPAddrs")
			if err != nil {
				b.Error(err)
				break
			}

			if writer != nil {
				writer.Stop()
			}

			writer, err = nsq.NewProducer(nsqdTCPAddrs, conf)
			if err != nil {
				b.Error(err)
				break
			}

		case msg := <-b.in:
			if writer == nil {
				continue
			}
			msgBytes, err := json.Marshal(msg)
			if err != nil {
				b.Error(err)
				break
			}
			if len(msgBytes) == 0 {
				continue
			}
			err = writer.Publish(topic, msgBytes)
			if err != nil {
				b.Error(err)
				break
			}

		case <-b.quit:
			if writer != nil {
				writer.Stop()
			}
			return
		case c := <-b.queryrule:
			c <- map[string]interface{}{
				"Topic":        topic,
				"NsqdTCPAddrs": nsqdTCPAddrs,
			}
		}
	}
}
