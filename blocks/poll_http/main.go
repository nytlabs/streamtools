package main

import (
	"flag"
	"github.com/bitly/go-simplejson"
	"github.com/nytlabs/streamtools/streamtools"
	"log"
)

var (
	pollInterval = flag.Float64("poll-interval", 300, "check S3 every X seconds")
	writeTopic   = flag.String("write-topic", "", "streamtools topic to write to")
	endpoint     = flag.String("endpoint", "", "http endpoint to poll using GET")
	rulePort     = flag.String("rule_port", "8080", "port to listen for new rules on")
	name         = flag.String("name", "poll-http", "name of block")
)

func main() {
	flag.Parse()
	streamtools.SetupLogger(name)
	block := streamtools.NewOutBlock(streamtools.PollHttp, *name)
	rule, err := simplejson.NewJson([]byte("{}"))
	if err != nil {
		log.Fatal(err.Error())
	}
	rule.Set("duration", *pollInterval)
	rule.Set("endpoint", *endpoint)
	block.RuleChan <- rule
	block.Run(*writeTopic, *rulePort)
}
