package library

import (
	"encoding/json"
	"errors"
	"github.com/nytlabs/gojee"                 // jee
	"github.com/nytlabs/streamtools/st/blocks" // blocks
	"io/ioutil"
	"net/http"
)

// specify those channels we're going to use to communicate with streamtools
type GetHTTP struct {
	blocks.Block
	queryrule chan chan interface{}
	inrule    chan interface{}
	in        chan interface{}
	out       chan interface{}
	quit      chan interface{}
}

// we need to build a simple factory so that streamtools can make new blocks of this kind
func NewGetHTTP() blocks.BlockInterface {
	return &GetHTTP{}
}

// Setup is called once before running the block. We build up the channels and specify what kind of block this is.
func (b *GetHTTP) Setup() {
	b.Kind = "GetHTTP"
	b.in = b.InRoute("in")
	b.inrule = b.InRoute("rule")
	b.queryrule = b.QueryRoute("rule")
	b.quit = b.Quit()
	b.out = b.Broadcast()
}

// Run is the block's main loop. Here we listen on the different channels we set up.
func (b *GetHTTP) Run() {
	client := &http.Client{}
	var tree *jee.TokenTree
	var path string
	for {
		select {
		case ruleI := <-b.inrule:
			// set a parameter of the block
			rule, ok := ruleI.(map[string]string)
			if !ok {
				b.Error(errors.New("could not assert rule to map[string]string"))
				continue
			}
			path, ok = rule["Path"]
			if !ok {
				b.Error(errors.New("could not find Path in rule"))
				continue
			}
			token, err := jee.Lexer(rule["Path"])
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
			urlInterface, err := jee.Eval(tree, msg)
			if err != nil {
				b.Error(err)
				continue
			}
			urlString, ok := urlInterface.(string)
			if !ok {
				b.Error(errors.New("couldn't assert url to a string"))
				continue
			}

			resp, err := client.Get(urlString)
			defer resp.Body.Close()
			if err != nil {
				b.Error(err)
				continue
			}

			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				b.Error(err)
				continue
			}
			var outMsg interface{}
			err = json.Unmarshal(body, &outMsg)
			if err != nil {
				b.Error(err)
				continue
			}
			b.out <- outMsg
		case respChan := <-b.queryrule:
			// deal with a query request
			respChan <- map[string]interface{}{
				"Path": path,
			}

		}
	}
}
