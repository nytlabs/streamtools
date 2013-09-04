package streamtools

import (
	"github.com/bitly/go-simplejson"
)

func DemuxByValue(inChan chan simplejson.Json, outChan chan simplejson.Json, RuleChan chan simplejson.Json) {

	rules := <-RuleChan

	key := rules.Get("key").String()

	for {
		select {
		case <-ruleChan:
		case msg := <-inChan:
			outTopic = msg.Get(key).String()

		}

	}

}
