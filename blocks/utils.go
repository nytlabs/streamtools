package blocks

import (
	"encoding/json"
	"log"
)

// broadcast emits a message to all output channels.
func broadcast(channels map[string]chan BMsg, msg BMsg) {
	for _, c := range channels {
		c <- msg
	}
}

// updateOutChans adds or deletes output channels for a block. updateOutChans
// is required for any block that has the ability to output to another block.
// It is used to add or delete another block's input channel to a block's
// output channels.
func updateOutChans(msg *OutChanMsg, b *Block) {
	switch msg.Action {
	case CREATE_OUT_CHAN:
		b.OutChans[msg.ID] = msg.OutChan
	case DELETE_OUT_CHAN:
		delete(b.OutChans, msg.ID)
	}
}

// unmarshal accepts a message that has been routed from the daemon, unmarshals
// it into JSON, and echoes the current value of the supplied struct back to
// the daemon. It is typically used to change state within a block from an HTTP
// handler.
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

// quit closes all input channels for a block.
func quit(b *Block) {
	close(b.InChan)
	for _, v := range b.Routes {
		close(v)
	}
	log.Println("quitting \"" + b.ID + "\" of type " + b.BlockType)
}
