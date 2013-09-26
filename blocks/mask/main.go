package main

import (
	"flag"
	"github.com/bitly/go-simplejson"
	"github.com/nytlabs/stream_tools/streamtools"
	"log"
)

var (
	readTopic  = flag.String("read_topic", "", "topic to write to")
	writeTopic = flag.String("write_topic", "", "topic to write to")
	command    = flag.String("mask", "", "mask JSON")
)

func main() {
	log.SetFlags(log.Lshortfile)
	flag.Parse()

	log.Println("reading from", *readTopic)
	log.Println("writing to", *writeTopic)
	log.Println("mask ", *command)

	block := streamtools.NewTransferBlock(streamtools.Mask, "mask")
	rule, err := simplejson.NewJson([]byte(*command))
	if err != nil {
		log.Fatal(err.Error())
	}
	block.RuleChan <- rule
	block.Run(*readTopic, *writeTopic, "8080")
}
