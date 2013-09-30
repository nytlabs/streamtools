package main

import (
	"flag"
	"github.com/bitly/go-simplejson"
	"github.com/nytlabs/streamtools/streamtools"
	"log"
)

var (
	readTopic   = flag.String("read-topic", "", "NSQ topic to read from")
	writeTopic  = flag.String("write-topic", "", "streamtools topic to write to")
	lookupdAddr = flag.String("lookupd-http-address", "127.0.0.1:4161", "lookupd address")
	maxInFlight = flag.Float64("max-in-flight", 1, "how many messages will be transferred at a time")
	name        = flag.String("name", "from-NSQ", "name of block")
)

func main() {
	flag.Parse()
	streamtools.SetupLogger(name)

	fromNSQBlock := streamtools.NewOutBlock(streamtools.FromNSQ, *name)
	rule, err := simplejson.NewJson([]byte("{}"))
	if err != nil {
		log.Fatal(err.Error())
	}
	rule.Set("readTopic", *readTopic)
	rule.Set("lookupdAddr", *lookupdAddr)
	rule.Set("maxInFlight", *maxInFlight)
	fromNSQBlock.RuleChan <- rule
	fromNSQBlock.Run(*writeTopic, "8080")
}
