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
)

func main() {
	log.SetFlags(log.Lshortfile)
	flag.Parse()
	block := streamtools.NewOutBlock(streamtools.MultiPollHttp, "httpPoller")
	rule, err := simplejson.NewJson([]byte("{}"))
	if err != nil {
		log.Fatal(err.Error())
	}
	rule.Set("duration", *pollInterval)
	rule.Set("endpoints", *endpoints)
	block.RuleChan <- rule
	block.Run(*writeTopic, *rulePort)
}
