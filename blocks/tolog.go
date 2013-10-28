package blocks

import (
	"log"
)

func ToLog(b *Block) {
	for {
		select {
		case msg := <-b.InChan:
			msgStr, err := msg.MarshalJSON()
			if err != nil {
				log.Println("wow bad json")
			}
			log.Println(string(msgStr))
		case msg := <-b.AddChan:
			updateOutChans(msg, b)
		case <-b.QuitChan:
			quit(b)
			return
		}
	}
}
