package library

import (
	"errors"

	"github.com/nytlabs/gojee"                 // jee
	"github.com/nytlabs/streamtools/st/blocks" // blocks
	"github.com/nytlabs/streamtools/st/util"   // util
)

// specify those channels we're going to use to communicate with streamtools
type DeDupe struct {
	blocks.Block
	queryrule chan blocks.MsgChan
	inrule    blocks.MsgChan
	in        blocks.MsgChan
	out       blocks.MsgChan
	quit      blocks.MsgChan
}

// we need to build a simple factory so that streamtools can make new blocks of this kind
func NewDeDupe() blocks.BlockInterface {
	return &DeDupe{}
}

// Setup is called once before running the block. We build up the channels and specify what kind of block this is.
func (b *DeDupe) Setup() {
	b.Kind = "Core"
	b.Desc = "stores a set of messages as specified by Path, emiting only those it hasn't seen before."
	b.in = b.InRoute("in")
	b.inrule = b.InRoute("rule")
	b.queryrule = b.QueryRoute("rule")
	b.quit = b.Quit()
	b.out = b.Broadcast()
}

// Run is the block's main loop. Here we listen on the different channels we set up.
func (b *DeDupe) Run() {
	var path string
	set := make(map[interface{}]bool)
	var tree *jee.TokenTree
	var err error
	for {
		select {
		case ruleI := <-b.inrule:
			// set a parameter of the block
			path, err = util.ParseString(ruleI, "Path")
			tree, err = util.BuildTokenTree(path)
			if err != nil {
				b.Error(err)
				break
			}
		case <-b.quit:
			// quit the block
			return
			// deal with inbound data
		case msg := <-b.in:
			if tree == nil {
				continue
			}
			v, err := jee.Eval(tree, msg)
			if err != nil {
				b.Error(err)
				break
			}

			if _, ok := v.(string); !ok {
				b.Error(errors.New("can only dedupe sets of strings"))
				continue
			}

			_, ok := set[v]
			// emit the incoming message if it isn't found in the set
			if !ok {
				b.out <- msg
				set[v] = true // and add it to the set
			}
		case c := <-b.queryrule:
			// deal with a query request
			c <- map[string]interface{}{
				"Path": path,
			}

		}
	}
}
