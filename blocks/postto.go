package blocks

import (
	"github.com/bitly/go-simplejson"
)

func PostTo(b *Block) {
	for {
		select {
		case msg := <-b.AddChan:
			updateOutChans(msg, b)
		case msg := <-b.Routes["in"]:
			outMsgJson, err := simplejson.NewJson(msg.Msg)
			if err != nil {
				msg.ResponseChan <- []byte(string(`{"Post":"Error"}`))
			} else {
				msg.ResponseChan <- []byte(string(`{"Post":"OK"}`))
				broadcast(b.OutChans, outMsgJson)
			}
		}
	}
}
