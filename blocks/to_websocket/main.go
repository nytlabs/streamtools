package main

import (
	"flag"
	"github.com/bitly/go-simplejson"
	"github.com/nytlabs/streamtools/streamtools"
)

var (
	topic  = flag.String("topic", "", "topic to read from")
	name   = flag.String("name", "to-websocket", "name of block")
	port   = flag.String("port", "8888", "port to serve websocket")
	handle = flag.String("handle", "/ws", "handle to serve websocket on eg: /ws")
)

func main() {
	flag.Parse()
	block := streamtools.NewInBlock(streamtools.ToWebSocket, *name)

	rule, _ := simplejson.NewJson([]byte("{}"))
	rule.Set("port", *port)
	rule.Set("handle", *handle)
	block.RuleChan <- rule

	block.Run(*topic, "8080")
}
