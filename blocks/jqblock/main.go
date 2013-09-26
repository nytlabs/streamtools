package main

import (
	"flag"
	"github.com/bitly/go-simplejson"
	"github.com/nytlabs/streamtools/streamtools"
	"log"
)

var (
	readTopic  = flag.String("read_topic", "", "topic to write to")
	writeTopic = flag.String("write_topic", "", "topic to write to")
	command    = flag.String("command", "", "jq command")
)

func main() {
	log.SetFlags(log.Lshortfile)
	flag.Parse()

	log.Println("reading from", *readTopic)
	log.Println("writing to", *writeTopic)
	log.Println("calling jq with command", *command)

	block := streamtools.NewTransferBlock(streamtools.JQ, "jq")
	rule, err := simplejson.NewJson([]byte("{}"))
	if err != nil {
		log.Fatal(err.Error())
	}
	rule.Set("command", *command)
	block.RuleChan <- rule
	block.Run(*readTopic, *writeTopic, "8080")
}
