package streamtools

import (
	"github.com/bitly/go-simplejson"
	"log"
)

func DemuxByValue(inChan chan simplejson.Json, outChan chan simplejson.Json, RuleChan chan simplejson.Json) {

	rules := <-RuleChan

	key, err := rules.Get("key").String()
	if err != nil {
		log.Fatal(err.Error())
	}

	for {
		select {
		case <-RuleChan:
		case msg := <-inChan:
			outTopic, err := msg.Get(key).String()
			if err != nil {
				log.Fatal(err.Error())
			}
			outMsg, err := simplejson.NewJson([]byte("{}"))
			if err != nil {
				log.Fatal(err.Error())
			}
			outMsg.Set("_StreamtoolsTopic", outTopic)
			outMsg.Set("_StreamtoolsData", msg)
			outChan <- *outMsg

		}

	}

}
