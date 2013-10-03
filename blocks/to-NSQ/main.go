package main

import (
	"flag"
	"github.com/bitly/go-simplejson"
	"github.com/nytlabs/streamtools/streamtools"
	"log"
)

var (
	nsqdHTTPAddrs = flag.String("nsqd-HTTP-address", "127.0.0.1:4161", "address of the NSQ daemon's HTTP address")
	writeTopic    = flag.String("write-topic", "", "topic to write to")
	readTopic     = flag.String("read-topic", "", "topic to read from")
	name          = flag.String("name", "to-NSQ", "name of block")
	rulePort      = flag.String("port", "8080", "port to listen for new rules on")
)

func main() {
	flag.Parse()
	streamtools.SetupLogger(name)
	block := streamtools.NewInBlock(streamtools.ToNSQ, *name)
	rule, err := simplejson.NewJson([]byte("{}"))
	if err != nil {
		log.Fatal(err.Error())
	}
	rule.Set("nsqdHTTPAddrs", *nsqdHTTPAddrs)
	rule.Set("writeTopic", *writeTopic)
	block.RuleChan <- rule
	block.Run(*readTopic, *rulePort)
}
