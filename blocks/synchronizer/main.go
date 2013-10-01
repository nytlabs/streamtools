package main

import (
	"flag"
	"github.com/bitly/go-simplejson"
	"github.com/nytlabs/streamtools/streamtools"
	"log"
)

var (
	lagTime    = flag.Float64("lag", 10, "duration of lag to synchronize in seconds")
	timeKey    = flag.String("key", "", "json key to use for time")
	readTopic  = flag.String("read-topic", "", "topic to write to")
	writeTopic = flag.String("write-topic", "", "topic to write to")
	name       = flag.String("name", "synchroniser", "name of block")
)

func main() {
	flag.Parse()
	streamtools.SetupLogger(name)
	block := streamtools.NewTransferBlock(streamtools.Synchronizer, *name)
	// make sure the block has a key waiting for it
	rule, err := simplejson.NewJson([]byte("{}"))
	if err != nil {
		log.Fatal(err.Error())
	}
	rule.Set("lag", *lagTime)
	rule.Set("key", *timeKey)
	block.RuleChan <- rule
	// set it going
	block.Run(*readTopic, *writeTopic, "8080")
}
