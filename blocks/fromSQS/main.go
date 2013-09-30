package main

import (
	"flag"
	"github.com/bitly/go-simplejson"
	"github.com/nytlabs/streamtools/streamtools"
	"log"
)

var (
	SQSEndpoint = flag.String("SQSEndpoint", "", "The SQS Endpoint you would like to listen to")
	writeTopic  = flag.String("write_topic", "", "streamtools topic to write to")
	name        = flag.String("name", "fromSQS", "name of block")
)

func main() {
	flag.Parse()
	streamtools.SetupLogger(name)
	SQSBlock := streamtools.NewOutBlock(streamtools.FromSQS, *name)
	rule, err := simplejson.NewJson([]byte("{}"))
	if err != nil {
		log.Fatal(err.Error())
	}
	rule.Set("SQSEndpoint", *SQSEndpoint)
	SQSBlock.RuleChan <- rule
	SQSBlock.Run(*writeTopic, "8080")

}
