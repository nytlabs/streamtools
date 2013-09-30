package main

import (
	"flag"
	"github.com/bitly/go-simplejson"
	"github.com/nytlabs/streamtools/streamtools"
	"log"
)

var (
	pollInterval = flag.Float64("poll-interval", 300, "poll every X seconds")
	writeTopic   = flag.String("write-topic", "", "streamtools topic to write to")
	endpoints    = flag.String("endpoint", "", `in the form [{"endpoint":"http://bla.com", "name":"bla"}]`)
	rulePort     = flag.String("rule_port", "8080", "port to listen for new rules on")
	name         = flag.String("name", "multi-poll-http", "name of block")
)

func main() {
	flag.Parse()
	streamtools.SetupLogger(name)
	block := streamtools.NewOutBlock(streamtools.MultiPollHttp, *name)
	rule, err := simplejson.NewJson([]byte("{}"))
	if err != nil {
		log.Fatal(err.Error())
	}
	rule.Set("duration", *pollInterval)
	rule.Set("endpoints", *endpoints)
	block.RuleChan <- rule
	block.Run(*writeTopic, *rulePort)
}
