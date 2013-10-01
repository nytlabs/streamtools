package main

import (
	"flag"
	"github.com/bitly/go-simplejson"
	"github.com/nytlabs/streamtools/streamtools"
	"log"
)

var (
	topic = flag.String("topic", "", "topic to read from")
	key   = flag.String("key", "", "key whose value you would like to find the minimum of")
	name  = flag.String("name", "min", "name of block")
)

func main() {
	flag.Parse()
	streamtools.SetupLogger(name)
	min := streamtools.NewStateBlock(streamtools.Min, *name)
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
