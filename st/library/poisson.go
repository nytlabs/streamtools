package library

import (
	"errors"
	"math"
	"math/rand"

	"github.com/nytlabs/streamtools/st/blocks" // blocks
	"github.com/nytlabs/streamtools/st/util"   // util
)

// specify those channels we're going to use to communicate with streamtools
type Poisson struct {
	blocks.Block
	queryrule chan blocks.MsgChan
	inrule    blocks.MsgChan
	inpoll    blocks.MsgChan
	out       blocks.MsgChan
	quit      blocks.MsgChan
}

// we need to build a simple factory so that streamtools can make new blocks of this kind
func NewPoisson() blocks.BlockInterface {
	return &Poisson{}
}

// Setup is called once before running the block. We build up the channels and specify what kind of block this is.
func (b *Poisson) Setup() {
	b.Kind = "Stats"
	b.Desc = "draws a random number from a Poisson distribution when polled"
	b.inrule = b.InRoute("rule")
	b.queryrule = b.QueryRoute("rule")
	b.inpoll = b.InRoute("poll")
	b.quit = b.Quit()
	b.out = b.Broadcast()
}

// algorithm due to Knuth http://en.wikipedia.org/wiki/Poisson_distribution
/*
algorithm poisson random number (Knuth):
    init:
         Let L ← e−^λ, k ← 0 and p ← 1.
    do:
         k ← k + 1.
         Generate uniform random number u in [0,1] and let p ← p × u.
    while p > L.
    return k − 1.
*/

func NewPoissonSampler(λ float64) func() int {
	L := math.Exp(-λ)
	r := rand.New(rand.NewSource(12345))
	return func() int {
		k := 0
		p := 1.0
		for {
			k = k + 1
			u := r.Float64()
			p = p * u
			if p <= L {
				return k - 1
			}
		}
	}
}

func (b *Poisson) Run() {
	var err error
	λ := 1.0
	sampler := NewPoissonSampler(λ)
	for {
		select {
		case ruleI := <-b.inrule:
			// set a parameter of the block
			rule, ok := ruleI.(map[string]interface{})
			if !ok {
				b.Error(errors.New("couldn't assert rule to map"))
			}
			λ, err = util.ParseFloat(rule, "Rate")
			if err != nil {
				b.Error(err)
			}
			sampler = NewPoissonSampler(λ)
		case <-b.quit:
			// quit the block
			return
		case <-b.inpoll:
			// deal with a poll request
			b.out <- map[string]interface{}{
				"sample": float64(sampler()),
			}
		case c := <-b.queryrule:
			// deal with a query request
			c <- map[string]interface{}{
				"Rate": λ,
			}
		}
	}
}
