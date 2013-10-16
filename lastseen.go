package streamtools

import (
	"fmt"
	"github.com/bitly/go-simplejson"
	"log"
)

type LastSeenBlock struct {
	AbstractBlock
}

func (b LastSeenBlock) blockRoutine() {
	log.Println("starting LastSeen block")
	lastSeen, _ := simplejson.NewJson([]byte("{}"))
	for {
		select {
		case msg := <-b.inChan:
			log.Println("lastseen got a message")
			lastSeen = msg
		case query := <-b.queryChan:
			log.Println("recieved query")
			fmt.Fprintln(query.w, lastSeen)
		}
	}
}

func NewLastSeen() Block {
	b := new(LastSeenBlock)
	b.blockType = "lastseen"
	b.ID = <-idChan
	return b
}
