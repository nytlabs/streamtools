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
			b.outChan <- msg
		case <-b.queryChan:
			log.Println("recieved query")
			log.Println(lastSeen)
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
	// note that whoever makes the connection must bless
	// the struct with channels before running it
	return b
}
