package blocks

import (
	"encoding/json"
)

// PostTo accepts JSON through POSTs to the /in endpoint and broadcasts to other blocks.
func PostTo(b *Block) {
	for {
		select {
		case msg := <-b.AddChan:
			updateOutChans(msg, b)
		case msg := <-b.Routes["in"]:
			var out BMsg
			err := json.Unmarshal(msg.Msg, &out)
			if err != nil {
				msg.ResponseChan <- []byte(string(`{"Post":"Error"}`))
			} else {
				msg.ResponseChan <- []byte(string(`{"Post":"OK"}`))
				broadcast(b.OutChans, out)
			}
		case <-b.QuitChan:
			quit(b)
			return
		}
	}
}
