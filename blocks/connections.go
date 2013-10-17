package blocks

import (
	"github.com/bitly/go-simplejson"
)

type Connection struct {
	AbstractBlock
}

func (b Connection) BlockRoutine() {
	lastSeen, _ := simplejson.NewJson([]byte("{}"))
	for {
		select {
		case msg := <-b.inChan:
			lastSeen = msg

			broadcast(b.outChans, msg)
		case query := <-b.routes["query"]:
			query.ResponseChan <- lastSeen
		}
	}
}

func NewConnection() Block {
	// create an empty ticker
	b := new(Connection)
	// specify the type for library
	b.blockType = "connection"
	// get the id
	//
	b.routes = map[string]chan RouteResponse{
		"query": make(chan RouteResponse),
	}
	// note that whoever makes the connection must bless
	// the struct with channels before running it
	return b
}
