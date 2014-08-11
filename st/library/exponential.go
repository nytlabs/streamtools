package library

import (
	"errors"
	"math/rand"

	"github.com/nytlabs/streamtools/st/blocks" // blocks
	"github.com/nytlabs/streamtools/st/util"   // util
)

type Exponential struct {
	blocks.Block
	queryrule chan blocks.MsgChan
	inrule    blocks.MsgChan
	inpoll    blocks.MsgChan
	out       blocks.MsgChan
	quit      blocks.MsgChan
}

func NewExponential() blocks.BlockInterface {
	return &Exponential{}
}

func (b *Exponential) Setup() {
	b.Kind = "Stats"
	b.Desc = "draws a random number from a Exponential distribution when polled"
	b.inrule = b.InRoute("rule")
	b.queryrule = b.QueryRoute("rule")
	b.inpoll = b.InRoute("poll")
	b.quit = b.Quit()
	b.out = b.Broadcast()
}

func (b *Exponential) Run() {
	var err error
	λ := 1.0
	for {
		select {
		case ruleI := <-b.inrule:
			// set a parameter of the block
			rule, ok := ruleI.(map[string]interface{})
			if !ok {
				b.Error(errors.New("couldn't assert rule to map"))
			}
			λ, err = util.ParseFloat(rule, "rate")
			if err != nil {
				b.Error(err)
			}
		case <-b.quit:
			// quit the block
			return
		case <-b.inpoll:
			// deal with a poll request
			b.out <- map[string]interface{}{
				"sample": rand.ExpFloat64(),
			}
		case c := <-b.queryrule:
			// deal with a query request
			c <- map[string]interface{}{
				"rate": λ,
			}
		}
	}
}
