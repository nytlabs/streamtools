package library

import (
	"errors"

	"github.com/nytlabs/gojee"                 // jee
	"github.com/nytlabs/streamtools/st/blocks" // blocks
	"github.com/nytlabs/streamtools/st/util"
)

// specify those channels we're going to use to communicate with streamtools
type Unpack struct {
	blocks.Block
	queryrule chan blocks.MsgChan
	inrule    blocks.MsgChan
	in        blocks.MsgChan
	out       blocks.MsgChan
	quit      blocks.MsgChan
}

// we need to build a simple factory so that streamtools can make new blocks of this kind
func NewUnpack() blocks.BlockInterface {
	return &Unpack{}
}

// Setup is called once before running the block. We build up the channels and specify what kind of block this is.
func (b *Unpack) Setup() {
	b.Kind = "Core"
	b.Desc = "splits an array of objects from incoming data, emitting each element as a separate message"
	b.in = b.InRoute("in")
	b.inrule = b.InRoute("rule")
	b.queryrule = b.QueryRoute("rule")
	b.quit = b.Quit()
	b.out = b.Broadcast()
}

// Run is the block's main loop. Here we listen on the different channels we set up.
func (b *Unpack) Run() {
	var path string
	var err error
	var tree *jee.TokenTree
	for {
		select {
		case ruleI := <-b.inrule:
			// set a parameter of the block
			rule, ok := ruleI.(map[string]interface{})
			if !ok {
				b.Error(errors.New("cannot assert rule to map"))
			}
			path, err = util.ParseString(rule, "Path")
			if err != nil {
				b.Error(err)
				continue
			}
			token, err := jee.Lexer(path)
			if err != nil {
				b.Error(err)
				continue
			}
			tree, err = jee.Parser(token)
			if err != nil {
				b.Error(err)
				continue
			}

		case <-b.quit:
			// quit the block
			return
		case msg := <-b.in:
			if tree == nil {
				continue
			}
			arrInterface, err := jee.Eval(tree, msg)
			if err != nil {
				b.Error(err)
				continue
			}
			arr, ok := arrInterface.([]interface{})
			if !ok {
				b.Error(errors.New("cannot assert " + path + " to array"))
				continue
			}
			for _, out := range arr {
				b.out <- out
			}
		case c := <-b.queryrule:
			// deal with a query request
			out := map[string]interface{}{
				"Path": path,
			}
			c <- out
		}
	}
}
