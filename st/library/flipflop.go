package library

import (
	"github.com/nytlabs/streamtools/st/blocks" // blocks
)

type FlipFlop struct {
	blocks.Block
	queryrule  chan chan interface{}
	querystate chan chan interface{}
	inrule     chan interface{}
	inpoll     chan interface{}
	in         chan interface{}
	out        chan interface{}
	quit       chan interface{}
}

// a bit of boilerplate for streamtools
func NewFlipFlop() blocks.BlockInterface {
	return &FlipFlop{}
}

func (b *FlipFlop) Setup() {
	b.Kind = "FlipFlop"
	b.in = b.InRoute("in")
	b.inpoll = b.InRoute("poll")
	b.querystate = b.QueryRoute("count")
	b.quit = b.Quit()
	b.out = b.Broadcast()
}

func (b *FlipFlop) Run() {

	state := false

	for {
		select {
		case <-b.quit:
			return
		case <-b.in:
			state = !state
		case <-b.inpoll:
			b.out <- map[string]interface{}{
				"state": state,
			}
		case c := <-b.querystate:
			c <- map[string]interface{}{
				"state": state,
			}
		}
	}
}
