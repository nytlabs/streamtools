package main

import (
	"flag"
	"github.com/bitly/go-simplejson"
	"github.com/nytlabs/streamtools/streamtools"
	"log"
)

var (
	readTopic  = flag.String("read-topic", "", "topic to read from")
	writeTopic = flag.String("write-topic", "", "topic to write to")
	key        = flag.String("key", "", "key agsinst which to filter")
	filter     = flag.String("filter", "", "key's value must equal this to proceed")
	name       = flag.String("name", "filter_by_key_value", "name of block")
)

func main() {
	flag.Parse()
	streamtools.SetupLogger(name)

	log.Println("reading from", *readTopic)
	log.Println("writing to", *writeTopic)
	log.Println("key ", *key)
	log.Println("filter ", *filter)

	block := streamtools.NewTransferBlock(streamtools.FilterByKeyValue, *name)
	rule, err := simplejson.NewJson([]byte("{}"))
	if err != nil {
		log.Fatal(err.Error())
	}
	rule.Set("key", *key)
	rule.Set("filter", *filter)
	block.RuleChan <- rule
	block.Run(*readTopic, *writeTopic, "8080")
}
