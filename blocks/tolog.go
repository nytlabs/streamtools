package blocks

import (
	"github.com/bitly/go-simplejson"
	"log"
)

type ToLogBlock struct {
	AbstractBlock
}

func (b ToLogBlock) BlockRoutine() {
	log.Println("starting to log block")
	for {
		select {
		case msg := <-b.inChan:
			msgStr, err := msg.MarshalJSON()
			if err != nil {
				log.Println("wow bad json")
			}
			log.Println(string(msgStr))
		}
	}
}

func NewToLog() Block {
	// create an empty ticker
	b := new(ToLogBlock)
	// specify the type for library
	b.blockType = "tolog"
	// start the inChan
	b.inChan = make(chan *simplejson.Json)
	return b
}
