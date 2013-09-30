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
	command    = flag.String("command", "", "jq command")
	name       = flag.String("name", "jq-block", "name of block")
)

func main() {
	flag.Parse()
	streamtools.SetupLogger(name)

	log.Println("reading from", *readTopic)
	log.Println("writing to", *writeTopic)
	log.Println("calling jq with command", *command)

	block := streamtools.NewTransferBlock(streamtools.JQ, *name)
	rule, err := simplejson.NewJson([]byte("{}"))
	if err != nil {
		log.Fatal(err.Error())
	}
	rule.Set("command", *command)
	block.RuleChan <- rule
	block.Run(*readTopic, *writeTopic, "8080")
}
