package main

import (
	"flag"
	"github.com/bitly/go-simplejson"
	"github.com/nytlabs/streamtools/streamtools"
	"log"
)

var (
	keymapping = flag.String("keymapping", "", "data key to query parameter mapping json")
	writeTopic = flag.String("write_topic", "", "topic to write to")
	readTopic  = flag.String("read_topic", "", "topic to write to")
	endpoint   = flag.String("endpoint", "", "endpoint")
)

func main() {
	log.SetFlags(log.Lshortfile)

	flag.Parse()

	log.Println("reading from", *readTopic)
	log.Println("writing to", *writeTopic)
	log.Println("using endpoint", *endpoint)

	block := streamtools.NewTransferBlock(streamtools.FetchJSON, "fetch_json")

	keymappingjson := `{"keymappings":` + *keymapping + "}"

	rule, err := simplejson.NewJson([]byte(keymappingjson))
	if err != nil {
		log.Fatal(err.Error())
	}
	rule.Set("endpoint", *endpoint)
	block.RuleChan <- rule
	block.Run(*readTopic, *writeTopic, "8080")
}
