package library

import (
	"errors"

	"github.com/nytlabs/gojee"                 // jee
	"github.com/nytlabs/streamtools/st/blocks" // blocks
	"github.com/nytlabs/streamtools/st/util"   // util
)

// specify those channels we're going to use to communicate with streamtools
type ToHTTPGetRequest struct {
	blocks.Block
	queryrule chan blocks.MsgChan
	inrule    blocks.MsgChan
	in        blocks.MsgChan
	quit      blocks.MsgChan
}

// we need to build a simple factory so that streamtools can make new blocks of this kind
func NewToHTTPGetRequest() blocks.BlockInterface {
	return &ToHTTPGetRequest{}
}

func (b *ToHTTPGetRequest) Setup() {
	b.Kind = "Network I/O"
	b.Desc = "responds to a Get requets's response channel"
	b.inrule = b.InRoute("rule")
	b.queryrule = b.QueryRoute("rule")
	b.quit = b.Quit()
	b.in = b.InRoute("in")
}

// Run is the block's main loop. Here we listen on the different channels we set up.
func (b *ToHTTPGetRequest) Run() {
	var respPath, msgPath string

	var respTree, msgTree *jee.TokenTree
	var err error

	for {
		select {
		case ruleI := <-b.inrule:
			respPath, err = util.ParseString(ruleI, "RespPath")
			if err != nil {
				b.Error(err)
				break
			}
			respTree, err = util.BuildTokenTree(respPath)
			if err != nil {
				b.Error(err)
				break
			}
			msgPath, err = util.ParseString(ruleI, "MsgPath")
			if err != nil {
				b.Error(err)
				break
			}
			msgTree, err = util.BuildTokenTree(msgPath)
			if err != nil {
				b.Error(err)
				break
			}
		case <-b.quit:
			return
		case msg := <-b.in:
			if respTree == nil {
				continue
			}
			if msgTree == nil {
				continue
			}
			cI, err := jee.Eval(respTree, msg)
			if err != nil {
				b.Error(err)
				break
			}
			c, ok := cI.(blocks.MsgChan)
			if !ok {
				b.Error(errors.New("response path must point to a channel"))
				continue
			}
			m, err := jee.Eval(msgTree, msg)
			if err != nil {
				b.Error(err)
				break
			}
			c <- m
		case responseChan := <-b.queryrule:
			// deal with a query request
			responseChan <- map[string]interface{}{
				"RespPath": respPath,
				"MsgPath":  msgPath,
			}
		}
	}
}
