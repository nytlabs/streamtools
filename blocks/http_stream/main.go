package main

import (
	"flag"
	"github.com/bitly/go-simplejson"
	"github.com/nytlabs/streamtools/streamtools"
	"log"
)

var (
	writeTopic = flag.String("write_topic", "", "topic to write to")
	endpoint   = flag.String("endpoint", "", "endpoint to listen")
	auth       = flag.String("auth", "", "optional usr:pwd string")
)

func main() {
	log.SetFlags(log.Lshortfile)
	flag.Parse()

	log.Println("[HTTPSTREAM] writing to", *writeTopic)
	log.Println("[HTTPSTREAM] using endpoint", *endpoint)

	block := streamtools.NewOutBlock(streamtools.FromHTTP, "fromHTTP")

	rule, err := simplejson.NewJson([]byte("{}"))
	if err != nil {
		log.Fatal(err.Error())
	}
	rule.Set("endpoint", *endpoint)
	rule.Set("auth", *auth)
	block.RuleChan <- rule

	block.Run(*writeTopic, "8080")
}
