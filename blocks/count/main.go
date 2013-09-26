package main

import (
	"flag"
	"github.com/bitly/go-simplejson"
	"github.com/nytlabs/stream_tools/streamtools"
	"log"
)

var (
	topic    = flag.String("topic", "", "topic to read from")
	window   = flag.Float64("window", 10, "size of window in seconds")
	rulePort = flag.String("port", "8080", "port to listen for new rules on")
)

func main() {
	log.SetFlags(log.Lshortfile)
	flag.Parse()
	block := streamtools.NewStateBlock(streamtools.Count, "count")
	// make sure the block has a key waiting for it
	rule, err := simplejson.NewJson([]byte("{}"))
	if err != nil {
		log.Fatal(err.Error())
	}
	rule.Set("window", *window)
	block.RuleChan <- rule
	// set it going
	block.Run(*topic, *rulePort)
}
