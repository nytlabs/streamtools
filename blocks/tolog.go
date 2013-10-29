package blocks

import (
	"log"
	"encoding/json"
)

func ToLog(b *Block) {
	for {
		select {
		case msg := <-b.InChan:
			out, err := json.Marshal(msg)
			if err != nil {
				log.Println("could not marshal json")
			}
			log.Println(string(out))
		case msg := <-b.AddChan:
			updateOutChans(msg, b)
		case <-b.QuitChan:
			quit(b)
			return
		}
	}
}
