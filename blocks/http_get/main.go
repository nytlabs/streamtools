package main

import (
	"flag"
	"github.com/bitly/go-simplejson"
	"github.com/nytlabs/streamtools/streamtools"
	"log"
)

var (
	keymapping = flag.String("keymapping", "", "data key to query parameter mapping json")
	readTopic  = flag.String("read_topic", "", "topic to write to")
	endpoint   = flag.String("endpoint", "", "endpoint")
)

func main() {
	log.SetFlags(log.Lshortfile)

	flag.Parse()

	log.Println("reading from", *readTopic)
	log.Println("using endpoint", *endpoint)

	block := streamtools.NewInBlock(streamtools.ToHTTP, "http_get")

	keymappingjson := `{"keymappings":` + *keymapping + "}"

	rule, err := simplejson.NewJson([]byte(keymappingjson))
	if err != nil {
		log.Fatal(err.Error())
	}
	rule.Set("endpoint", *endpoint)
	block.RuleChan <- rule
	block.Run(*readTopic, "8080")
}
