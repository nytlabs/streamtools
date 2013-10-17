package streamtools

import (
	"github.com/bitly/go-simplejson"
	"log"
)

type Connection struct {
	AbstractBlock
}

func (b Connection) blockRoutine() {
	lastSeen, _ := simplejson.NewJson([]byte("{}"))
	for {
		select {
		case msg := <-b.inChan:
			lastSeen = msg
			log.Println(b.outChans)

			broadcast(b.outChans, msg)
		case query := <-b.routes["query"]:
			query.responseChan <- lastSeen
		}
	}
}

func NewConnection() Block {
	// create an empty ticker
	b := new(Connection)
	// specify the type for library
	b.blockType = "connection"
	// get the id
	b.ID = <-idChan
	//
	b.routes = map[string]chan routeResponse{
		"query": make(chan routeResponse),
	}
	// note that whoever makes the connection must bless
	// the struct with channels before running it
	return b
}
