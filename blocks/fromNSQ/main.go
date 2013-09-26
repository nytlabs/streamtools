package main

import (
	"flag"
	"github.com/bitly/go-simplejson"
	"github.com/nytlabs/streamtools/streamtools"
	"log"
)

var (
	readTopic   = flag.String("read_topic", "", "NSQ topic to read from")
	writeTopic  = flag.String("write_topic", "", "streamtools topic to write to")
	lookupdAddr = flag.String("lookupd-http-address", "127.0.0.1:4161", "lookupd address")
)

func main() {
	flag.Parse()
	fromNSQBlock := streamtools.NewOutBlock(streamtools.FromNSQ, "fromNSQ")
	rule, err := simplejson.NewJson([]byte("{}"))
	if err != nil {
		log.Fatal(err.Error())
	}
	rule.Set("readTopic", *readTopic)
	rule.Set("lookupdAddr", *lookupdAddr)
	fromNSQBlock.RuleChan <- rule
	fromNSQBlock.Run(*writeTopic, "8080")
}
