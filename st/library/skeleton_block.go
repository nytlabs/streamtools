package library

import (
	"github.com/nytlabs/streamtools/st/blocks" // blocks
)

// specify those channels we're going to use to communicate with streamtools
type Skeleton struct {
	blocks.Block
	queryrule chan blocks.MsgChan
	inrule    blocks.MsgChan
	inpoll    blocks.MsgChan
	in        blocks.MsgChan
	out       blocks.MsgChan
	quit      blocks.MsgChan
}

// we need to build a simple factory so that streamtools can make new blocks of this kind
func NewSkeleton() blocks.BlockInterface {
	return &Skeleton{}
}

// Setup is called once before running the block. We build up the channels and specify what kind of block this is.
func (b *Skeleton) Setup() {
	b.Kind = "Skeleton"
	b.Desc = "use this block as a starting template for creating new blocks"
	b.in = b.InRoute("in")
	b.inrule = b.InRoute("rule")
	b.queryrule = b.QueryRoute("rule")
	b.inpoll = b.InRoute("poll")
	b.quit = b.Quit()
	b.out = b.Broadcast()
}

// Run is the block's main loop. Here we listen on the different channels we set up.
func (b *Skeleton) Run() {
	for {
		select {
		case ruleI := <-b.inrule:
			// set a parameter of the block
			_, _ = ruleI.(map[string]interface{})
		case <-b.quit:
			// quit the block
			return
		case _ = <-b.in:
			// deal with inbound data
		case <-b.inpoll:
			// deal with a poll request
		case _ = <-b.queryrule:
			// deal with a query request
		}
	}
}
