package blocks

import (
	"github.com/bitly/go-simplejson"
    "encoding/json"
    "log"
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

func unmarshal(r RouteResponse, rule interface{}){
    json.Unmarshal([]byte(r.Msg), &rule)
    m, err := json.Marshal(rule)
    if err != nil{
        log.Println("could not unmarshal rule")
    }
    r.ResponseChan <- string(m)
}

func marshal(r RouteResponse, rule interface{}){
    m, err := json.Marshal(rule)
    if err != nil{
        log.Println("could not marshal rule")
    }
    r.ResponseChan <- string(m)
}