package streamtools

import (
	"github.com/bitly/go-simplejson"
	"log"
)

// TODO have the ability to filter on things that aren't strings
// TODO have the ability to invert the filter

func FilterByKeyValue(inChan chan *simplejson.Json, outChan chan *simplejson.Json, RuleChan chan *simplejson.Json) {
	rules := <-RuleChan
	key, err := rules.Get("key").String()
	if err != nil {
		log.Fatal(err)
	}
	filter, err := rules.Get("filter").String()
	if err != nil {
		log.Fatal(err)
	}

	for {
		select {

		case <-RuleChan:
		case msg := <-inChan:

			val, err := getKey(key, msg).String()
			if err != nil {
				log.Fatal(err)
			}
			log.Println(val)
			log.Println(filter)
			if val == filter {
				outChan <- msg
			}
		}
	}
}
