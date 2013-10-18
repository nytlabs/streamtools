package blocks

import (
	"github.com/bitly/go-simplejson"
)

func Connection(b *BlockDefinition) {
	lastSeen, _ := simplejson.NewJson([]byte("{}"))
	for {
		select {
		case msg := <-b.InChan:
			lastSeen = msg
			broadcast(b.OutChans, msg)
		case query := <-b.Routes["query"]:
			query.ResponseChan <- lastSeen
		}
	}
}
