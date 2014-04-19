package library

import (
	"encoding/json"
	"github.com/bitly/go-nsq"
	"github.com/nytlabs/streamtools/st/blocks" // blocks
	"github.com/nytlabs/streamtools/st/util"
	"time"
)

// specify those channels we're going to use to communicate with streamtools
type ToNSQMulti struct {
	blocks.Block
	queryrule    chan chan interface{}
	inrule       chan interface{}
	in           chan interface{}
	out          chan interface{}
	quit         chan interface{}
	nsqdTCPAddrs string
	topic        string
}

// a bit of boilerplate for streamtools
func NewToNSQMulti() blocks.BlockInterface {
	return &ToNSQMulti{}
}

func (b *ToNSQMulti) Setup() {
	b.Kind = "ToNSQMulti"
	b.Desc = "sends messages to an NSQ topic in batches"
	b.in = b.InRoute("in")
	b.inrule = b.InRoute("rule")
	b.queryrule = b.QueryRoute("rule")
	b.quit = b.Quit()
}

// connects to an NSQ topic and emits each message into streamtools.
func (b *ToNSQMulti) Run() {
	var writer *nsq.Writer
	var batch [][]byte
	interval := time.Duration(1 * time.Second)
	maxBatch := 100

	dump := time.NewTicker(interval)
	for {
		select {
		case <-dump.C:
			if writer == nil || len(batch) == 0 {
				break
			}
			_, _, err := writer.MultiPublish(b.topic, batch)
			if err != nil {
				b.Error(err.Error())
			}

			batch = nil
		case ruleI := <-b.inrule:
			//rule := ruleI.(map[string]interface{})

			topic, err := util.ParseString(ruleI, "Topic")
			if err != nil {
				b.Error(err)
				break
			}

			nsqdTCPAddrs, err := util.ParseString(ruleI, "NsqdTCPAddrs")
			if err != nil {
				b.Error(err)
				break
			}

			intervalS, err := util.ParseString(ruleI, "Interval")
			if err != nil {
				b.Error("bad input")
				break
			}

			dur, err := time.ParseDuration(intervalS)
			if err != nil {
				b.Error(err)
				break
			}

			if dur <= 0 {
				b.Error("interval must be positive")
				break
			}

			batchSize, err := util.ParseFloat(ruleI, "MaxBatch")
			if err != nil {
				b.Error("error parsing batch size")
				break
			}

			if writer != nil {
				writer.Stop()
			}

			maxBatch = int(batchSize)
			interval = dur

			dump.Stop()
			dump = time.NewTicker(interval)
			writer = nsq.NewWriter(nsqdTCPAddrs)
			b.topic = topic
			b.nsqdTCPAddrs = nsqdTCPAddrs
		case msg := <-b.in:
			if writer == nil {
				break
			}

			msgByte, err := json.Marshal(msg)
			if err != nil {
				b.Error(err)
			}
			batch = append(batch, msgByte)

			if len(batch) > maxBatch {
				_, _, err := writer.MultiPublish(b.topic, batch)
				if err != nil {
					b.Error(err.Error())
					break
				}
				batch = nil
			}
		case <-b.quit:
			if writer != nil {
				writer.Stop()
			}
			dump.Stop()
			return
		case c := <-b.queryrule:
			c <- map[string]interface{}{
				"Topic":        b.topic,
				"NsqdTCPAddrs": b.nsqdTCPAddrs,
				"MaxBatch":     maxBatch,
				"Interval":     interval.String(),
			}
		}
	}
}
