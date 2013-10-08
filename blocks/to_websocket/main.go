package main

import (
	"flag"
	"github.com/bitly/go-simplejson"
	"github.com/nytlabs/streamtools/streamtools"
)

var (
	topic  = flag.String("topic", "", "topic to read from")
	name   = flag.String("name", "to-websocket", "name of block")
	port   = flag.String("port", "8080", "rule port")
	wsport   = flag.String("ws-port", "8888", "port to serve websocket")
	wshandle = flag.String("ws-handle", "/ws", "handle to serve websocket on eg: /ws")
)

func main() {
	flag.Parse()
	block := streamtools.NewInBlock(streamtools.ToWebSocket, *name)

	rule, _ := simplejson.NewJson([]byte("{}"))
	rule.Set("wsport", *wsport)
	rule.Set("wshandle", *wshandle)
	block.RuleChan <- rule

	block.Run(*topic, *port)
}
