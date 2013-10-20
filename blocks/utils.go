package blocks

import (
	"github.com/bitly/go-simplejson"
)

func broadcast(channels map[string]chan *simplejson.Json, msg *simplejson.Json) {
	for _, c := range channels {
		c <- msg
	}
}

func updateOutChans(msg *OutChanMsg, b *Block) {
	switch msg.Action {
	case CREATE_OUT_CHAN:
		b.OutChans[msg.ID] = msg.OutChan
	}
}
