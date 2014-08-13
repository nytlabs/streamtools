package library

import (
	elastigo "github.com/mattbaird/elastigo/lib"
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

// a bit of boilerplate for streamtools
func NewToElasticsearch() blocks.BlockInterface {
	return &ToElasticsearch{}
}

func (b *ToElasticsearch) Setup() {
	b.Kind = "Data Stores"
	b.Desc = "sends messages as JSON to a specified index and type in Elasticsearch"
	b.in = b.InRoute("in")
	b.inrule = b.InRoute("rule")
	b.queryrule = b.QueryRoute("rule")
	b.quit = b.Quit()
}

// connects to an NSQ topic and emits each message into streamtools.
func (b *ToElasticsearch) Run() {
	var err error
	var esIndex string
	var esType string

	conn := elastigo.NewConn()

	host := "localhost"
	port := "9200"

	for {
		select {
		case ruleI := <-b.inrule:
			host, err = util.ParseString(ruleI, "Host")
			if err != nil {
				b.Error(err)
				break
			}

			port, err = util.ParseString(ruleI, "Port")
			if err != nil {
				b.Error(err)
				break
			}

			esIndex, err = util.ParseString(ruleI, "Index")
			if err != nil {
				b.Error(err)
				break
			}

			esType, err = util.ParseString(ruleI, "Type")
			if err != nil {
				b.Error(err)
				break
			}

			conn.Domain = host
			conn.Port = port

		case msg := <-b.in:
			_, err := conn.Index(esIndex, esType, "", nil, msg)
			if err != nil {
				b.Error(err)
			}
		case <-b.quit:
			return
		case c := <-b.queryrule:
			c <- map[string]interface{}{
				"Host":  host,
				"Port":  port,
				"Index": esIndex,
				"Type":  esType,
			}
		}
	}
}
