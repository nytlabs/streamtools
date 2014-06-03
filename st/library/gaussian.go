package library

import (
	"errors"
	"math/rand"

	"github.com/nytlabs/streamtools/st/blocks" // blocks
	"github.com/nytlabs/streamtools/st/util"   // util
)

// specify those channels we're going to use to communicate with streamtools
type Gaussian struct {
	blocks.Block
	queryrule chan blocks.MsgChan
	inrule    blocks.MsgChan
	inpoll    blocks.MsgChan
	out       blocks.MsgChan
	quit      blocks.MsgChan
}

// we need to build a simple factory so that streamtools can make new blocks of this kind
func NewGaussian() blocks.BlockInterface {
	return &Gaussian{}
}

// Setup is called once before running the block. We build up the channels and specify what kind of block this is.
func (b *Gaussian) Setup() {
	b.Kind = "Stats"
	b.Desc = "draws a random number from the Gaussian distribution when polled"
	b.inrule = b.InRoute("rule")
	b.queryrule = b.QueryRoute("rule")
	b.inpoll = b.InRoute("poll")
	b.quit = b.Quit()
	b.out = b.Broadcast()
}

// Run is the block's main loop. Here we listen on the different channels we set up.
func (b *Gaussian) Run() {
	var err error
	mean := 0.0
	stddev := 1.0

	for {
		select {
		case ruleI := <-b.inrule:
			// set a parameter of the block
			rule, ok := ruleI.(map[string]interface{})
			if !ok {
				b.Error(errors.New("couldn't assert rule to map"))
			}
			mean, err = util.ParseFloat(rule, "Mean")
			if err != nil {
				b.Error(err)
			}
			stddev, err = util.ParseFloat(rule, "StdDev")
			if err != nil {
				b.Error(err)
			}
		case <-b.quit:
			// quit the block
			return
		case <-b.inpoll:
			// deal with a poll request
			b.out <- map[string]interface{}{
				"sample": rand.NormFloat64()*stddev + mean,
			}
		case MsgChan := <-b.queryrule:
			// deal with a query request
			out := map[string]interface{}{
				"Mean":   mean,
				"StdDev": stddev,
			}
			MsgChan <- out
		}
	}
}
