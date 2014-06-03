package library

import (
	"github.com/nytlabs/streamtools/st/blocks" // blocks
)

type Toggle struct {
	blocks.Block
	in   blocks.MsgChan
	out  blocks.MsgChan
	quit blocks.MsgChan
}

// a bit of boilerplate for streamtools
func NewToggle() blocks.BlockInterface {
	return &Toggle{}
}

func (b *Toggle) Setup() {
	b.Kind = "Core"
	b.Desc = "emits a 'state' boolean value, toggling true/false on each hit"
	b.in = b.InRoute("in")
	b.quit = b.Quit()
	b.out = b.Broadcast()
}

func (b *Toggle) Run() {

	state := false

	for {
		select {
		case <-b.quit:
			return
		case <-b.in:
			state = !state
			b.out <- map[string]interface{}{
				"state": state,
			}
		}
	}
}
