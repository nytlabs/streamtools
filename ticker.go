package streamtools

import (
	"github.com/bitly/go-simplejson"
	"log"
	"time"
)

type TickerBlock struct {
	AbstractBlock
}

func (b TickerBlock) blockRoutine() {
	log.Println("starting Ticker block")
	ticker := time.NewTicker(time.Duration(2) * time.Second)
	outMsg, _ := simplejson.NewJson([]byte("{}"))
	for {
		select {
		case tick := <-ticker.C:
			outMsg.Set("t", tick)
			log.Println("hello")
			b.outChan <- outMsg
		}
	}
}

func NewTicker() Block {
	id := <-idChan

	b := new(TickerBlock)
	b.blockType = "ticker"
	b.ID = id
	return b
}
