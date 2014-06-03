package library

import (
	"time"

	"github.com/nytlabs/streamtools/st/blocks" // blocks
	"github.com/nytlabs/streamtools/st/util"
)

// specify those channels we're going to use to communicate with streamtools
type Ticker struct {
	blocks.Block
	queryrule chan blocks.MsgChan
	inrule    blocks.MsgChan
	out       blocks.MsgChan
	quit      blocks.MsgChan
}

// we need to build a simple factory so that streamtools can make new blocks of this kind
func NewTicker() blocks.BlockInterface {
	return &Ticker{}
}

// Setup is called once before running the block. We build up the channels and specify what kind of block this is.
func (b *Ticker) Setup() {
	b.Kind = "Core"
	b.Desc = "emits the time at an interval specified by the block's rule"
	b.inrule = b.InRoute("rule")
	b.queryrule = b.QueryRoute("rule")
	b.quit = b.Quit()
	b.out = b.Broadcast()
}

// Run is the block's main loop. Here we listen on the different channels we set up.
func (b *Ticker) Run() {
	interval := time.Duration(1) * time.Second
	ticker := time.NewTicker(interval)
	for {
		select {
		case tick := <-ticker.C:
			b.out <- map[string]interface{}{
				"tick": tick.String(),
			}
		case ruleI := <-b.inrule:
			// set a parameter of the block
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

			interval = dur
			ticker.Stop()
			ticker = time.NewTicker(interval)
		case <-b.quit:
			return
		case c := <-b.queryrule:
			// deal with a query request
			c <- map[string]interface{}{
				"Interval": interval.String(),
			}
		}
	}
}
