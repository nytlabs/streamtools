package blocks

import (
	"encoding/json"
	"log"
)

func broadcast(channels map[string]chan BMsg, msg BMsg) {
	for _, c := range channels {
		c <- msg
	}
}

func updateOutChans(msg *OutChanMsg, b *Block) {
	switch msg.Action {
	case CREATE_OUT_CHAN:
		b.OutChans[msg.ID] = msg.OutChan
	case DELETE_OUT_CHAN:
		delete(b.OutChans, msg.ID)
	}
}

func unmarshal(r RouteResponse, rule interface{}) {
	err := json.Unmarshal(r.Msg, &rule)
	if err != nil {
		log.Println("found errors during unmarshalling")
		log.Println(err.Error())
	}
	m, err := json.Marshal(rule)
	if err != nil {
		log.Println("could not marshal rule")
	}
	r.ResponseChan <- m
}

func marshal(r RouteResponse, rule interface{}) {
	m, err := json.Marshal(rule)
	if err != nil {
		log.Println("could not marshal rule")
	}
	r.ResponseChan <- m
}

func quit(b *Block) {
	close(b.InChan)
	for _, v := range b.Routes {
		close(v)
	}
	log.Println("quitting \"" + b.ID + "\" of type " + b.BlockType)
}
