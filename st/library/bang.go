package library

import (
	"github.com/nytlabs/streamtools/st/blocks" // blocks
)

// specify those channels we're going to use to communicate with streamtools
type Bang struct {
	blocks.Block
	query chan blocks.MsgChan
	out   blocks.MsgChan
	quit  blocks.MsgChan
}

// we need to build a simple factory so that streamtools can make new blocks of this kind
func NewBang() blocks.BlockInterface {
	return &Bang{}
}

// Setup is called once before running the block. We build up the channels and specify what kind of block this is.
func (b *Bang) Setup() {
	b.Kind = "Core"
	b.Desc = "sends a 'bang' request to blocks connected to it, triggered by clicking the query endpoint"
	b.query = b.QueryRoute("query")
	b.quit = b.Quit()
	b.out = b.Broadcast()
}

// Run is the block's main loop. Here we listen on the different channels we set up.
func (b *Bang) Run() {
	for {
		select {
		case <-b.quit:
			// quit the block
			return
		case c := <-b.query:
			// deal with inbound data
			out := map[string]interface{}{
				"Bang": "!",
			}
			c <- map[string]interface{}{
				"Bang": "!",
			}
			b.out <- out
		}
	}
}
