package blocks

import (
	"github.com/bitly/go-simplejson"
	"log"
	"time"
)

func Ticker(b *Block) {
	log.Println("starting Ticker block")
	ticker := time.NewTicker(time.Duration(2) * time.Second)
	for {
		select {
		case tick := <-ticker.C:
			outMsg, _ := simplejson.NewJson([]byte("{}"))
			outMsg.Set("t", tick)
			for _, oc := range b.OutChans {
				oc <- outMsg
			}
		case msg := <- b.AddChan:
			switch msg.Action {
			case CREATE_OUT_CHAN:
				b.OutChans[msg.ID] = msg.OutChan
			}
		}
	}
}
