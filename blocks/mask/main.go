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
	command    = flag.String("mask", "", "mask JSON")
	name       = flag.String("name", "mask", "name of block")
)

func main() {
	flag.Parse()
	streamtools.SetupLogger(name)

	log.Println("reading from", *readTopic)
	log.Println("writing to", *writeTopic)
	log.Println("mask ", *command)

	block := streamtools.NewTransferBlock(streamtools.Mask, *name)
	rule, err := simplejson.NewJson([]byte(*command))
	if err != nil {
		log.Fatal(err.Error())
	}
	block.RuleChan <- rule
	block.Run(*readTopic, *writeTopic, "8080")
}
