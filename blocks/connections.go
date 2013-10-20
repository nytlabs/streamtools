package blocks

import (
	"github.com/bitly/go-simplejson"
)

func Connection(b *Block) {
	lastSeen, _ := simplejson.NewJson([]byte("{}"))
	for {
		select {
		case msg := <-b.InChan:
			lastSeen = msg
			broadcast(b.OutChans, msg)
		case query := <-b.Routes["last_seen"]:
			query.ResponseChan <- lastSeen
		case msg := <- b.AddChan:
			switch msg.Action {
			case CREATE_OUT_CHAN:
				b.OutChans[msg.ID] = msg.OutChan
			}
		}
	}
}
