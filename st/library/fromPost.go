package library

import (
	"github.com/nytlabs/streamtools/st/blocks" // blocks
)

// specify those channels we're going to use to communicate with streamtools
type FromPost struct {
	blocks.Block
	queryrule chan blocks.MsgChan
	inrule    blocks.MsgChan
	in        blocks.MsgChan
	out       blocks.MsgChan
	quit      blocks.MsgChan
}

// we need to build a simple factory so that streamtools can make new blocks of this kind
func NewFromPost() blocks.BlockInterface {
	return &FromPost{}
}

// Setup is called once before running the block. We build up the channels and specify what kind of block this is.
func (b *FromPost) Setup() {
	b.Kind = "Network I/O"
	b.Desc = "emits any message that is POSTed to its IN route"
	b.in = b.InRoute("in")
	b.quit = b.Quit()
	b.out = b.Broadcast()
}

// Run is the block's main loop. Here we listen on the different channels we set up.
func (b *FromPost) Run() {
	for {
		select {
		case <-b.quit:
			// quit the block
			return
		case msg := <-b.in:
			// deal with inbound data
			b.out <- msg
		}
	}
}
