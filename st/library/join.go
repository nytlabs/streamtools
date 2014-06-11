package library

import (
	"github.com/nytlabs/streamtools/st/blocks" // blocks
)

type Join struct {
	blocks.Block
	inA   blocks.MsgChan
	inB   blocks.MsgChan
	clear blocks.MsgChan
	out   blocks.MsgChan
	quit  blocks.MsgChan
}

func NewJoin() blocks.BlockInterface {
	return &Join{}
}

func (b *Join) Setup() {
	b.Kind = "Core"
	b.Desc = "joins two streams together, emitting the joined message once it's been seen on both inputs"
	b.inA = b.InRoute("inA")
	b.inB = b.InRoute("inB")
	b.clear = b.InRoute("clear")
	b.quit = b.Quit()
	b.out = b.Broadcast()
}

func (b *Join) Run() {
	A := make(blocks.MsgChan, 1000)
	B := make(blocks.MsgChan, 1000)
	for {
		select {
		case <-b.quit:
			return
		case msg := <-b.inA:
			select {
			case A <- msg:
			default:
				b.Error("the A queue is overflowing")
			}
		case msg := <-b.inB:
			select {
			case B <- msg:
			default:
				b.Error("the B queue is overflowing")
			}
		case <-b.clear:
		Clear:
			for {
				select {
				case <-A:
				case <-B:
				default:
					break Clear
				}
			}
		}
		for len(A) > 0 && len(B) > 0 {
			b.out <- map[string]interface{}{
				"A": <-A,
				"B": <-B,
			}
		}
	}
}
