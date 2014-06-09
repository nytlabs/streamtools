package library

import (
	"errors"

	"github.com/nytlabs/gojee"                 // jee
	"github.com/nytlabs/streamtools/st/blocks" // blocks
	"github.com/nytlabs/streamtools/st/util"   // util
)

// specify those channels we're going to use to communicate with streamtools
type Set struct {
	blocks.Block
	queryrule   chan blocks.MsgChan
	inrule      blocks.MsgChan
	add         blocks.MsgChan
	isMember    blocks.MsgChan
	cardinality chan blocks.MsgChan
	out         blocks.MsgChan
	quit        blocks.MsgChan
}

// we need to build a simple factory so that streamtools can make new blocks of this kind
func NewSet() blocks.BlockInterface {
	return &Set{}
}

// Setup is called once before running the block. We build up the channels and specify what kind of block this is.
func (b *Set) Setup() {
	b.Kind = "Core"
	b.Desc = "add, ismember and cardinality routes on a stored set of values"

	// set operations
	b.add = b.InRoute("add")
	b.isMember = b.InRoute("isMember")
	b.cardinality = b.QueryRoute("cardinality")

	b.inrule = b.InRoute("rule")
	b.queryrule = b.QueryRoute("rule")
	b.quit = b.Quit()
	b.out = b.Broadcast()
}

// Run is the block's main loop. Here we listen on the different channels we set up.
func (b *Set) Run() {
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
		case msg := <-b.add:
			if tree == nil {
				continue
			}
			v, err := jee.Eval(tree, msg)
			if err != nil {
				b.Error(err)
				break
			}
			if _, ok := v.(string); !ok {
				b.Error(errors.New("can only build sets of strings"))
				continue
			}
			set[v] = true
			// deal with inbound data
		case msg := <-b.isMember:
			if tree == nil {
				continue
			}
			v, err := jee.Eval(tree, msg)
			if err != nil {
				b.Error(err)
				break
			}
			_, ok := set[v]
			b.out <- map[string]interface{}{
				"isMember": ok,
			}
		case c := <-b.cardinality:
			c <- map[string]interface{}{
				"cardinality": len(set),
			}
		case c := <-b.queryrule:
			// deal with a query request
			c <- map[string]interface{}{
				"Path": path,
			}

		}
	}
}
