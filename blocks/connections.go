package blocks

import (
	"github.com/bitly/go-simplejson"
)

func Connection(b *Block) {
	var last *simplejson.Json

	for {
		select {
		case msg := <-b.InChan:
			last = msg
			broadcast(b.OutChans, msg)
		case query := <-b.Routes["last_seen"]:
			r, _ := last.MarshalJSON()
			query.ResponseChan <- r
		case msg := <-b.AddChan:
			updateOutChans(msg, b)
		}
	}
}
