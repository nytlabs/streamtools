package blocks

import (
	"log"
)

func ToLog(b *Block) {
	log.Println("starting to log block")
	for {
		select {
		case msg := <-b.InChan:
			msgStr, err := msg.MarshalJSON()
			if err != nil {
				log.Println("wow bad json")
			}
			log.Println(string(msgStr))
		case msg := <- b.AddChan:
			switch msg.Action {
			case CREATE_OUT_CHAN:
				b.OutChans[msg.ID] = msg.OutChan
			}
		}
	}
}
