package main

import (
	"flag"
	"github.com/bitly/go-simplejson"
	"github.com/mikedewar/stream_tools/streamtools"
	"log"
)

var (
	pollInterval = flag.Float64("poll-interval", 300, "check S3 every X seconds")
	writeTopic   = flag.String("write-topic", "", "streamtools topic to write to")
	endpoint     = flag.String("endpoint", "", "http endpoint to poll using GET")
	rulePort     = flag.String("rule_port", "8080", "port to listen for new rules on")
)

func main() {
	log.SetFlags(log.Lshortfile)
	flag.Parse()
	block := streamtools.NewOutBlock(streamtools.PollHttp, "httpPoller")
	rule, err := simplejson.NewJson([]byte("{}"))
	if err != nil {
		log.Fatal(err.Error())
	}
	rule.Set("duration", *pollInterval)
	rule.Set("endpoint", *endpoint)
	block.RuleChan <- rule
	block.Run(*writeTopic, *rulePort)
}
