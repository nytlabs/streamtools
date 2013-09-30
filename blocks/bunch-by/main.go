package main

import (
	"flag"
	"github.com/bitly/go-simplejson"
	"github.com/nytlabs/streamtools/streamtools"
	"log"
)

var (
	key        = flag.String("key", "", "create bunches according to this key")
	after      = flag.Float64("after", 1800, "number of seconds to wait for no activity before emitting")
	readTopic  = flag.String("read-topic", "", "topic to read from")
	writeTopic = flag.String("write-topic", "", "topic to write to")
	name       = flag.String("name", "bunch-by", "name of block")
)

func main() {

	streamtools.SetupLogger(name)

	flag.Parse()

	block := streamtools.NewTransferBlock(streamtools.Bunch, "bunch")
	rule, err := simplejson.NewJson([]byte("{}"))
	if err != nil {
		log.Fatal(err.Error())
	}
	rule.Set("branch", *key)
	rule.Set("after", *after)
	block.RuleChan <- rule
	block.Run(*readTopic, *writeTopic, "8080")
}
