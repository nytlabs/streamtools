package blocks

import (
	"encoding/json"
	"log"
)

// ToLog prints recieved messages to the stream tools logger.
func ToLog(b *Block) {
	for {
		select {
		case msg := <-b.InChan:
			out, err := json.Marshal(msg)
			if err != nil {
				log.Println("could not marshal json")
			}
			log.Println(string(out))
		case <-b.QuitChan:
			quit(b)
			return
		}
	}
}
