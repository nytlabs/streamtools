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
	queryrule chan blocks.MsgChan
	inrule    blocks.MsgChan
	in        blocks.MsgChan
	out       blocks.MsgChan
	quit      blocks.MsgChan
}

// we need to build a simple factory so that streamtools can make new blocks of this kind
func NewToElasticsearch() blocks.BlockInterface {
	return &ToElasticsearch{}
}

// Setup is called once before running the block. We build up the channels and specify what kind of block this is.
func (b *ToElasticsearch) Setup() {
	b.Kind = "Data Stores"
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
	var err error
	var index string
	var indextype string

	host := "localhost"
	port := "9200"

	for {
		select {
		case msgI := <-b.inrule:
			host, err = util.ParseString(msgI, "Host")
			if err != nil {
				b.Error(err)
				continue
			}
			port, err = util.ParseString(msgI, "Port")
			if err != nil {
				b.Error(err)
				continue
			}
			index, err = util.ParseString(msgI, "Index")
			if err != nil {
				b.Error(err)
				continue
			}
			indextype, err = util.ParseString(msgI, "IndexType")
			if err != nil {
				b.Error(err)
				continue
			}

			// Set the Elasticsearch Host/Port to Connect to
			api.Domain = host
			api.Port = port

		case MsgChan := <-b.queryrule:
			// deal with a query request
			MsgChan <- map[string]interface{}{
				"Host":      host,
				"Port":      port,
				"Index":     index,
				"IndexType": indextype,
			}
		case <-b.quit:
			// quit the block
			return
		case msg := <-b.in:
			var args map[string]interface{}
			_, err := core.Index(index, indextype, "", args, msg)
			if err != nil {
				b.Error(err)
			}
		}
	}
}
