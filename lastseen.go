package streamtools

import (
	"github.com/bitly/go-simplejson"
	"log"
)

type LastSeenBlock struct {
	AbstractBlock
}

func (b LastSeenBlock) blockRoutine() {
	log.Println("starting block")
	lastSeen, _ := simplejson.NewJson([]byte("{}"))
	for {
		select {
		case msg := <-b.inChan:
			lastSeen = msg
		case query := <-b.queryChan:
			log.Println("recieved query")
			query.responseChan <- lastSeen
		}
	}
}

func NewLastSeen() Block {
	// create an empty block
	b := new(LastSeenBlock)
	// set the type
	b.blockType = "lastseen"
	// set the id
	b.ID = <-idChan
	// set the queryChan
	b.queryChan = make(chan query)
	// set the inChan
	b.inChan = make(chan *simplejson.Json)
	return b
}
