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
			out, err := lastSeen.MarshalJSON()
			if err != nil {
				log.Println(err.Error())
			}
			log.Println(string(out))
			_, err = fmt.Fprintf(query.w, string(out))
			if err != nil {
				log.Println(err.Error())
			}
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
