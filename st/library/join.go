package library

import (
	"github.com/nytlabs/streamtools/st/blocks" // blocks
)

type Join struct {
	blocks.Block
	inA   chan interface{}
	inB   chan interface{}
	clear chan interface{}
	out   chan interface{}
	quit  chan interface{}
}

func NewJoin() blocks.BlockInterface {
	return &Join{}
}

func (b *Join) Setup() {
	b.Kind = "Join"
	b.inA = b.InRoute("inA")
	b.inB = b.InRoute("inB")
	b.clear = b.InRoute("clear")
	b.quit = b.Quit()
	b.out = b.Broadcast()
}

func (b *Join) Run() {
	A := make(chan interface{}, 1000)
	B := make(chan interface{}, 1000)
	for {
		select {
		case <-b.quit:
			return
		case msg := <-b.inA:
			A <- msg
		case msg := <-b.inB:
			B <- msg
		case <-b.clear:
			go func() {
				for {
					select {
					case <-A:
					case <-B:
					default:
						return
					}
				}
			}()
		}
		for len(A) > 0 && len(B) > 0 {
			b.out <- map[string]interface{}{
				"A": <-A,
				"B": <-B,
			}
		}
	}
}
