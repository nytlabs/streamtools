package library

import (
	"fmt"
	"github.com/clbanning/mxj"
	"github.com/nytlabs/gojee"                 // jee
	"github.com/nytlabs/streamtools/st/blocks" // blocks
	"github.com/nytlabs/streamtools/st/util"
)

// specify those channels we're going to use to communicate with streamtools
type ParseXML struct {
	blocks.Block
	queryrule chan chan interface{}
	inrule    chan interface{}
	in        chan interface{}
	out       chan interface{}
	quit      chan interface{}
}

// we need to build a simple factory so that streamtools can make new blocks of this kind
func NewParseXML() blocks.BlockInterface {
	return &ParseXML{}
}

// Setup is called once before running the block. We build up the channels and specify what kind of block this is.
func (b *ParseXML) Setup() {
	b.Kind = "ParseXML"
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

			msi := map[string]interface{}(mapVal)
			for k, v := range msi {
				switch vt := v.(type) {
				case map[string]interface{}:
					fmt.Println("value at key", k, "is a msi", vt)
				case mxj.Map:
					fmt.Println("value at key", k, "is a mxj.Map", vt)
				default:
					fmt.Println("value at key", k, "is something else", vt)
				}

				fmt.Println("\t\t", k, ":", v)
			}

			b.out <- mapVal

		case respChan := <-b.queryrule:
			// deal with a query request
			respChan <- map[string]interface{}{
				"Path": path,
			}

		}
	}
}
