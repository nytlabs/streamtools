package blocks

import (
	"github.com/bitly/go-simplejson"
	"time"
)

func Ticker(b *Block) {
	ticker := time.NewTicker(time.Duration(2) * time.Second)
	for {
		select {
		case tick := <-ticker.C:
			outMsg, _ := simplejson.NewJson([]byte("{}"))
			outMsg.Set("t", tick)
			broadcast(b.OutChans, outMsg)
		case msg := <- b.AddChan:
			updateOutChans(msg, b)
		}
	}
}
