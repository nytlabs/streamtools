package blocks

import (
	"github.com/bitly/go-simplejson"
	"log"
	"time"
)

func Ticker(b *BlockDefinition) {
	log.Println("starting Ticker block")
	ticker := time.NewTicker(time.Duration(2) * time.Second)
	outMsg, _ := simplejson.NewJson([]byte("{}"))
	for {
		select {
		case tick := <-ticker.C:
			outMsg.Set("t", tick)
			for _, oc := range b.OutChans {
				oc <- outMsg
			}
		}
	}
}
