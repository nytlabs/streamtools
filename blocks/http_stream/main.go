package main

import (
	"flag"
	"github.com/bitly/go-simplejson"
	"github.com/nytlabs/streamtools/streamtools"
	"log"
)

var (
	writeTopic = flag.String("write-topic", "", "topic to write to")
	endpoint   = flag.String("endpoint", "", "endpoint to listen")
	auth       = flag.String("auth", "", "optional usr:pwd string")
	name       = flag.String("name", "from-http", "name of block")
)

func main() {
	flag.Parse()
	streamtools.SetupLogger(name)

	log.Println("writing to", *writeTopic)
	log.Println("using endpoint", *endpoint)

	block := streamtools.NewOutBlock(streamtools.FromHTTP, *name)

	rule, err := simplejson.NewJson([]byte("{}"))
	if err != nil {
		log.Fatal(err.Error())
	}
	rule.Set("endpoint", *endpoint)
	rule.Set("auth", *auth)
	block.RuleChan <- rule

	block.Run(*writeTopic, "8080")
}
