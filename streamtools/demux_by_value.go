package streamtools

import (
	"github.com/bitly/go-simplejson"
	"log"
	"strings"
)

func cleanTopicName(topic string) string {
	// TODO
	// Valid topic and channel names are characters [.a-zA-Z0-9_-] and 1 < length <= 32
	topic = strings.Replace(topic, ":", "-", -1)
	return topic
}

func DeMuxByValue(inChan chan *simplejson.Json, outChan chan *simplejson.Json, RuleChan chan *simplejson.Json) {

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
			outTopic = cleanTopicName(outTopic)
			outMsg, err := simplejson.NewJson([]byte("{}"))
			if err != nil {
				log.Fatal(err.Error())
			}
			outMsg.Set("_StreamtoolsTopic", outTopic)
			msgMap, err := msg.Map()
			if err != nil {
				log.Fatal(err.Error())
			}
			outMsg.Set("_StreamtoolsData", msgMap)
			outChan <- outMsg
		}

	}

}
