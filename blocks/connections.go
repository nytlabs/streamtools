package blocks

import (
	"encoding/json"
	"log"
)

func Connection(b *Block) {
	var last BMsg

	for {
		select {
		case msg := <-b.InChan:
			last = msg
			broadcast(b.OutChans, msg)
		case query := <-b.Routes["last_seen"]:
			mj, err := json.Marshal(last)
			if err != nil {
				log.Println(err.Error())
			}
			query.ResponseChan <- mj
		case msg := <-b.AddChan:
			updateOutChans(msg, b)
		case <-b.QuitChan:
			quit(b)
			return
		}
	}

}
