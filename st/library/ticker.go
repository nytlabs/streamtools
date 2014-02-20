package library

import (
	"github.com/nytlabs/streamtools/st/blocks" // blocks
	"time"
)

// specify those channels we're going to use to communicate with streamtools
type Ticker struct {
	blocks.Block
	queryrule chan chan interface{}
	inrule    chan interface{}
	out       chan interface{}
	quit      chan interface{}
}

// we need to build a simple factory so that streamtools can make new blocks of this kind
func NewTicker() blocks.BlockInterface {
	return &Ticker{}
}

// Setup is called once before running the block. We build up the channels and specify what kind of block this is.
func (b *Ticker) Setup() {
	b.Kind = "Ticker"
	b.inrule = b.InRoute("rule")
	b.queryrule = b.QueryRoute("rule")
	b.quit = b.InRoute("quit")
	b.out = b.Broadcast()
}

// Run is the block's main loop. Here we listen on the different channels we set up.
func (b *Ticker) Run() {
	var err error
	interval := time.Duration(1) * time.Second
	ticker := time.NewTicker(interval)
	for {
		select {
		case tick := <-ticker.C:
			b.out <- map[string]interface{}{
				"tick": tick,
			}
		case ruleI := <-b.inrule:
			// set a parameter of the block
			rule := ruleI.(map[string]string)
			interval, err = time.ParseDuration(rule["Interval"])
			if err != nil {
				b.Error(err)
			}
			ticker = time.NewTicker(interval)
		case <-b.quit:
			// quit the block
			return
		case c := <-b.queryrule:
			// deal with a query request
			c <- map[string]interface{}{
				"Interval": interval,
			}
		}
	}
}
