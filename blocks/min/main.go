package main

import (
	"flag"
	"github.com/bitly/go-simplejson"
	"github.com/mikedewar/stream_tools/streamtools"
	"log"
)

var (
	topic = flag.String("topic", "", "topic to read from")
	key   = flag.String("key", "", "key whose value you would like to find the minimum of")
)

func main() {
	log.SetFlags(log.Lshortfile)
	flag.Parse()
	min := streamtools.NewStateBlock(streamtools.Min, "min")
	// make sure the block has a key waiting for it
	rule, err := simplejson.NewJson([]byte("{}"))
	if err != nil {
		log.Fatal(err.Error())
	}
	rule.Set("key", *key)
	min.RuleChan <- rule
	// set it going
	min.Run(*topic, "8080")
}
