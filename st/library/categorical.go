package library

import (
	"errors"
	"math/rand"

	"github.com/nytlabs/streamtools/st/blocks" // blocks
	"github.com/nytlabs/streamtools/st/util"   // util
)

// specify those channels we're going to use to communicate with streamtools
type Categorical struct {
	blocks.Block
	queryrule chan blocks.MsgChan
	inrule    blocks.MsgChan
	inpoll    blocks.MsgChan
	out       blocks.MsgChan
	quit      blocks.MsgChan
}

// we need to build a simple factory so that streamtools can make new blocks of this kind
func NewCategorical() blocks.BlockInterface {
	return &Categorical{}
}

// Setup is called once before running the block. We build up the channels and specify what kind of block this is.
func (b *Categorical) Setup() {
	b.Kind = "Stats"
	b.Desc = "draws a random number from a Categorical distribution when polled"
	b.inrule = b.InRoute("rule")
	b.queryrule = b.QueryRoute("rule")
	b.inpoll = b.InRoute("poll")
	b.quit = b.Quit()
	b.out = b.Broadcast()
}

func NewCategoricalSampler(θ []float64) func() int {
	L := make([]float64, len(θ))
	for i, θi := range θ {
		if i == 0 {
			L[i] = θi
		} else {
			L[i] = L[i-1] + θi
		}
	}
	r := rand.New(rand.NewSource(12345))
	return func() int {
		u := r.Float64()
		for i, Li := range L {
			if Li >= u {
				return i
			}
		}
		return len(L) - 1 // this should never happen
	}
}

func (b *Categorical) Run() {
	var err error
	θ := []float64{1.0}
	sampler := NewCategoricalSampler(θ)
	for {
		select {
		case ruleI := <-b.inrule:
			// set a parameter of the block
			rule, ok := ruleI.(map[string]interface{})
			if !ok {
				b.Error(errors.New("couldn't assert rule to map"))
			}
			θ, err = util.ParseArrayFloat(rule, "Weights")
			if err != nil {
				b.Error(err)
			}
			// normalise!
			Z := 0.0
			for _, θi := range θ {
				Z += θi
			}
			if Z == 0 {
				b.Error(errors.New("Weights must not sum to zero"))
				continue
			}
			for i := range θ {
				θ[i] /= Z
			}

			sampler = NewCategoricalSampler(θ)
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
				"Weights": θ,
			}
		}
	}
}
