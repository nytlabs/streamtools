package library

import (
	"errors"

	"github.com/nytlabs/gojee"                 // jee
	"github.com/nytlabs/streamtools/st/blocks" // blocks
	"github.com/nytlabs/streamtools/st/util"   // util
)

// specify those channels we're going to use to communicate with streamtools
type LinearModel struct {
	blocks.Block
	queryrule chan blocks.MsgChan
	inrule    blocks.MsgChan
	in        blocks.MsgChan
	out       blocks.MsgChan
	quit      blocks.MsgChan
}

// we need to build a simple factory so that streamtools can make new blocks of this kind
func NewLinearModel() blocks.BlockInterface {
	return &LinearModel{}
}

// Setup is called once before running the block. We build up the channels and specify what kind of block this is.
func (b *LinearModel) Setup() {
	b.Kind = "Stats"
	b.Desc = "Emits the linear combination of paramters and features"
	b.inrule = b.InRoute("rule")
	b.queryrule = b.QueryRoute("rule")
	b.in = b.InRoute("in")
	b.quit = b.Quit()
	b.out = b.Broadcast()
}

// Run is the block's main loop. Here we listen on the different channels we set up.
func (b *LinearModel) Run() {

	var β []float64
	var featurePaths []string
	var featureTrees []*jee.TokenTree
	var err error

	for {
	Loop:
		select {
		case rule := <-b.inrule:
			β, err = util.ParseArrayFloat(rule, "Weights")
			if err != nil {
				b.Error(err)
				continue
			}
			featurePaths, err = util.ParseArrayString(rule, "FeaturePaths")
			if err != nil {
				b.Error(err)
				continue
			}
			featureTrees = make([]*jee.TokenTree, len(featurePaths))
			for i, path := range featurePaths {
				token, err := jee.Lexer(path)
				if err != nil {
					b.Error(err)
					break
				}
				tree, err := jee.Parser(token)
				if err != nil {
					b.Error(err)
					break
				}
				featureTrees[i] = tree
			}
		case <-b.quit:
			// quit the block
			return
		case msg := <-b.in:
			if featureTrees == nil {
				break
			}
			x := make([]float64, len(featureTrees))
			for i, tree := range featureTrees {
				feature, err := jee.Eval(tree, msg)
				if err != nil {
					b.Error(err)
					break Loop
				}
				fi, ok := feature.(float64)
				if !ok {
					b.Error(errors.New("features must be float64"))
					break Loop
				}
				x[i] = fi
			}
			y := 0.0
			for i, βi := range β {
				y += βi * x[i]
			}
			b.out <- map[string]interface{}{
				"Response": y,
			}

		case MsgChan := <-b.queryrule:
			out := map[string]interface{}{
				"Weights":      β,
				"FeaturePaths": featurePaths,
			}
			MsgChan <- out
		}
	}
}
