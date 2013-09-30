package main

import (
	"flag"
	"github.com/bitly/go-simplejson"
	"github.com/nytlabs/streamtools/streamtools"
	"log"
)

var (
	readTopic = flag.String("read-topic", "", "topic to read from")
	key       = flag.String("key", "", "key to demux on")
	name      = flag.String("name", "demux-by-value", "name of block")
)

func main() {

	flag.Parse()
	streamtools.SetupLogger(name)

	deMuxByValueBlock := streamtools.NewTransferBlock(streamtools.DeMuxByValue, "demux_by_value")

	rule, err := simplejson.NewJson([]byte("{}"))
	if err != nil {
		log.Fatal(err.Error())
	}
	rule.Set("key", *key)
	deMuxByValueBlock.RuleChan <- rule
	deMuxByValueBlock.DeMuxRun(*readTopic, "8080")
}
