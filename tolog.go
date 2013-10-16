package streamtools

import (
    "github.com/bitly/go-simplejson"
    "log"
)

type ToLogBlock struct {
    AbstractBlock
}

func (b ToLogBlock) blockRoutine() {
    log.Println("starting to log block")
    for {
        select {
        case msg := <- b.inChan:
            log.Println(msg)
        }
    }
}

func NewToLog() Block {
    // create an empty ticker
    b := new(ToLogBlock)
    // specify the type for library
    b.blockType = "tolog"
    // get the id
    b.ID = <-idChan
    // make the outChan
    b.inChan = make(chan *simplejson.Json)
    b.outChan = make(chan *simplejson.Json)
    return b
}
