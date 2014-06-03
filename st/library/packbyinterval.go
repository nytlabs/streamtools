package library

import (
	"time"

	"github.com/nytlabs/streamtools/st/blocks" // blocks
	"github.com/nytlabs/streamtools/st/util"
)

// specify those channels we're going to use to communicate with streamtools
type PackByInterval struct {
	blocks.Block
	queryrule chan blocks.MsgChan
	inrule    blocks.MsgChan
	clear     blocks.MsgChan
	flush     blocks.MsgChan
	in        blocks.MsgChan
	out       blocks.MsgChan
	quit      blocks.MsgChan
}

// we need to build a simple factory so that streamtools can make new blocks of this kind
func NewPackByInterval() blocks.BlockInterface {
	return &PackByInterval{}
}

// Setup is called once before running the block. We build up the channels and specify what kind of block this is.
func (b *PackByInterval) Setup() {
	b.Kind = "Core"
	b.Desc = "Packs incoming messages into array. Arrays are emitted and emptied on an interval."
	b.in = b.InRoute("in")
	b.inrule = b.InRoute("rule")
	b.queryrule = b.QueryRoute("rule")
	b.clear = b.InRoute("clear")
	b.flush = b.InRoute("flush")
	b.quit = b.Quit()
	b.out = b.Broadcast()
}

// Run is the block's main loop. Here we listen on the different channels we set up.
func (b *PackByInterval) Run() {
	var batch []interface{}

	interval := time.Duration(1) * time.Second
	ticker := time.NewTicker(interval)
	for {
		select {
		case <-ticker.C:
			b.out <- map[string]interface{}{
				"Pack": batch,
			}
			batch = nil

		case ruleI := <-b.inrule:
			intervalS, err := util.ParseString(ruleI, "Interval")
			if err != nil {
				b.Error("error parsing batch size")
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

			interval = dur
			ticker.Stop()
			ticker = time.NewTicker(interval)
			batch = nil
		case <-b.quit:
			// quit the block
			return
		case m := <-b.in:
			batch = append(batch, m)
		case <-b.clear:
			batch = nil
		case <-b.flush:
			b.out <- map[string]interface{}{
				"Pack": batch,
			}
			batch = nil
		case r := <-b.queryrule:
			r <- map[string]interface{}{
				"Interval": interval.String(),
			}
		}
	}
}
