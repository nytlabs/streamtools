package streamtools

import (
	"github.com/bitly/go-simplejson"
	"log"
	"strconv"
	"time"
)

type RouteExample struct {
	AbstractBlock
}

func (b RouteExample) blockRoutine() {
	period := 1000

	log.Println("starting Route Example block")
	ticker := time.NewTicker(time.Duration(period) * time.Millisecond)
	outMsg, _ := simplejson.NewJson([]byte("{}"))
	for {
		select {
		case tick := <-ticker.C:
			outMsg.Set("t", tick)
			broadcast(b.outChans, outMsg)
		case routeResp := <-b.routes["getRule"]:
			outMsg, _ := simplejson.NewJson([]byte(`{"period":` + strconv.Itoa(period) + `}`))
			routeResp.responseChan <- outMsg
		case routeResp := <-b.routes["setRule"]:
			p, err := routeResp.msg.Get("period").Int()
			if err != nil {
				respMsg, _ := simplejson.NewJson([]byte(`{"status":"NOT OK"}`))
				routeResp.responseChan <- respMsg
				break
			}

			period = p
			ticker = time.NewTicker(time.Duration(period) * time.Millisecond)

			respMsg, _ := simplejson.NewJson([]byte(`{"status":"OK"}`))
			routeResp.responseChan <- respMsg

		case routeResp := <-b.routes["hello"]:
			respMsg, _ := simplejson.NewJson([]byte(`{"HELLO":"WORLD"}`))
			routeResp.responseChan <- respMsg
		case routeResp := <-b.routes["writeMsg"]:
			w, _ := routeResp.msg.Get("message").String()
			log.Println(w)
			respMsg, _ := simplejson.NewJson([]byte(`{"status":"OK"}`))
			routeResp.responseChan <- respMsg
		}
	}
}

func NewRouteExample() Block {
	// create an empty ticker
	b := new(RouteExample)
	// specify the type for library
	b.blockType = "RouteExample"
	//routes
	b.routes = map[string]chan routeResponse{
		"setRule":  make(chan routeResponse),
		"getRule":  make(chan routeResponse),
		"hello":    make(chan routeResponse),
		"writeMsg": make(chan routeResponse),
	}
	return b
}
