package library

import (
	"github.com/nytlabs/streamtools/st/blocks" // blocks
)

// specify those channels we're going to use to communicate with streamtools
type FromHTTPGetRequest struct {
	blocks.Block
	query chan blocks.MsgChan
	out   blocks.MsgChan
	quit  blocks.MsgChan
}

// we need to build a simple factory so that streamtools can make new blocks of this kind
func NewFromHTTPGetRequest() blocks.BlockInterface {
	return &FromHTTPGetRequest{}
}

// Setup is called once before running the block. We build up the channels and specify what kind of block this is.
func (b *FromHTTPGetRequest) Setup() {
	b.Kind = "Network I/O"
	b.Desc = "emits a query route that must be responded to using another block"
	b.query = b.QueryRoute("query")
	b.quit = b.Quit()
	b.out = b.Broadcast()
}

// Run is the block's main loop. Here we listen on the different channels we set up.
func (b *FromHTTPGetRequest) Run() {
	for {
		select {
		case <-b.quit:
			// quit the block
			return
		case c := <-b.query:
			// deal with inbound data
			out := map[string]interface{}{
				"MsgChan": c,
			}
			b.out <- out
		}
	}
}
