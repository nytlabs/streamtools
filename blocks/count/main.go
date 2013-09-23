package main

import (
	"flag"
	"github.com/bitly/go-simplejson"
	"github.com/mikedewar/stream_tools/streamtools"
	"log"
)

var (
	topic  = flag.String("topic", "", "topic to read from")
	window = flag.Float64("window", 10, "size of window in seconds")
)

func main() {
	log.SetFlags(log.Lshortfile)
	flag.Parse()
	min := streamtools.NewStateBlock(streamtools.Count, "min")
	// make sure the block has a key waiting for it
	rule, err := simplejson.NewJson([]byte("{}"))
	if err != nil {
		log.Fatal(err.Error())
	}
	rule.Set("window", *window)
	min.RuleChan <- rule
	// set it going
	min.Run(*topic, "8080")
}
