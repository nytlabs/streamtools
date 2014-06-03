package library

import (
	"encoding/json"

	"github.com/nytlabs/gojee" // jee
	"github.com/nytlabs/mxj"
	"github.com/nytlabs/streamtools/st/blocks" // blocks
	"github.com/nytlabs/streamtools/st/util"
)

// specify those channels we're going to use to communicate with streamtools
type ParseXML struct {
	blocks.Block
	queryrule chan blocks.MsgChan
	inrule    blocks.MsgChan
	in        blocks.MsgChan
	out       blocks.MsgChan
	quit      blocks.MsgChan
}

// we need to build a simple factory so that streamtools can make new blocks of this kind
func NewParseXML() blocks.BlockInterface {
	return &ParseXML{}
}

// Setup is called once before running the block. We build up the channels and specify what kind of block this is.
func (b *ParseXML) Setup() {
	b.Kind = "Parsers"
	b.Desc = "converts incoming XML messages to JSON for use in streamtools"
	b.in = b.InRoute("in")
	b.inrule = b.InRoute("rule")
	b.queryrule = b.QueryRoute("rule")
	b.quit = b.Quit()
	b.out = b.Broadcast()
}

// Run is the block's main loop. Here we listen on the different channels we set up.
func (b *ParseXML) Run() {
	var tree *jee.TokenTree
	var path string
	var err error
	var xmlData []byte

	for {
		select {
		case ruleI := <-b.inrule:
			// set a parameter of the block
			path, err = util.ParseString(ruleI, "Path")
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
			// deal with inbound data
			if tree == nil {
				continue
			}
			dataI, err := jee.Eval(tree, msg)
			if err != nil {
				b.Error(err)
				continue
			}

			switch v := dataI.(type) {
			case []byte:
				xmlData = v

			case string:
				xmlData = []byte(v)

			default:
				b.Error("data should be a string or a []byte")
				continue
			}

			// parse xml -> map[string]interface{} with mxj
			// http://godoc.org/github.com/clbanning/mxj#NewMapXml
			mapVal, err := mxj.NewMapXml(xmlData)
			if err != nil {
				b.Error(err)
				continue
			}

			// TODO: replace this json.Marshal / Unmarshal dance with Nik's recursive map copy from the map block
			outMsg, err := json.Marshal(mapVal)
			if err != nil {
				b.Error(err)
				continue
			}

			var newMsg interface{}
			err = json.Unmarshal(outMsg, &newMsg)
			if err != nil {
				b.Error(err)
				continue
			}

			b.out <- newMsg

		case MsgChan := <-b.queryrule:
			// deal with a query request
			MsgChan <- map[string]interface{}{
				"Path": path,
			}

		}
	}
}
