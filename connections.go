package streamtools

import (
	"fmt"
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
		case query := <-b.queryChan:
			log.Println("recieved query")
			fmt.Fprintln(query.w, lastSeen)
		}
	}
}
