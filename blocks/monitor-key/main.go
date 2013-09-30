package main

import (
	"flag"
	"github.com/bitly/go-simplejson"
	"github.com/nytlabs/streamtools/streamtools"
	"log"
)

var (
	key       = flag.String("key", "", "create bunches according to this key")
	readTopic = flag.String("read-topic", "", "topic to read from")
	port      = flag.String("port", "8080", "port to listen for new rules on")
	name      = flag.String("name", "monitor-key", "name of block")
)

func main() {
	flag.Parse()
	streamtools.SetupLogger(name)
	block := streamtools.NewStateBlock(streamtools.MonitorKey, *name)
	rule, err := simplejson.NewJson([]byte("{}"))
	if err != nil {
		log.Fatal(err.Error())
	}
	rule.Set("key", *key)
	block.RuleChan <- rule
	block.Run(*readTopic, *port)
}
