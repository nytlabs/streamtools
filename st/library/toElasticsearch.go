package library

import (
	"github.com/mattbaird/elastigo/api"
	"github.com/mattbaird/elastigo/core"
	"github.com/nytlabs/streamtools/st/blocks" // blocks
	"github.com/nytlabs/streamtools/st/util"
)

// specify those channels we're going to use to communicate with streamtools
type ToElasticsearch struct {
	blocks.Block
	queryrule chan chan interface{}
	inrule    chan interface{}
	in        chan interface{}
	out       chan interface{}
	quit      chan interface{}
	host      string
	port      string
	index     string
	indextype string
}

// we need to build a simple factory so that streamtools can make new blocks of this kind
func NewToElasticsearch() blocks.BlockInterface {
	return &ToElasticsearch{}
}

// Setup is called once before running the block. We build up the channels and specify what kind of block this is.
func (b *ToElasticsearch) Setup() {
	b.Kind = "ToElasticsearch"
	b.Desc = "sends messages as JSON to a specified index and type in Elasticsearch"
	b.in = b.InRoute("in")
	b.inrule = b.InRoute("rule")
	b.queryrule = b.QueryRoute("rule")
	b.quit = b.Quit()
	b.out = b.Broadcast()
}

// Run is the block's main loop. Here we listen on the different channels we set up.
// This block posts a message to a specified Elasticsearch index with the given type.
func (b *ToElasticsearch) Run() {
	for {
		select {
		case msgI := <-b.inrule:
			host, _ := util.ParseString(msgI, "Host")
			port, _ := util.ParseString(msgI, "Port")
			index, _ := util.ParseString(msgI, "Index")
			indextype, _ := util.ParseString(msgI, "IndexType")

			// Set the Elasticsearch Host/Port to Connect to
			api.Domain = host
			api.Port = port

			b.host = host
			b.port = port
			b.index = index
			b.indextype = indextype

		case respChan := <-b.queryrule:
			// deal with a query request
			respChan <- map[string]interface{}{
				"Host":      b.host,
				"Port":      b.port,
				"Index":     b.index,
				"IndexType": b.indextype,
			}
		case <-b.quit:
			// quit the block
			return
		case msg := <-b.in:
			var args map[string]interface{}
			_, err := core.Index(b.index, b.indextype, "", args, msg)
			if err != nil {
				b.Error(err)
			}
		}
	}
}
