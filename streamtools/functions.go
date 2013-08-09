package streamtools

import (
	"github.com/bitly/go-simplejson"
)

type TransferFunction func(inChan chan simplejson.Json, outChan chan simplejson.Json)

func TransferBlock(inTopic string, outTopic string, channel string, f TransferFunction) {
	ex := make(chan bool)
	inChan := make(chan simplejson.Json)
	outChan := make(chan simplejson.Json)
	go nsqReader(inTopic, channel, inChan)
	go f(inChan, outChan)
	go nsqWriter(outTopic, outChan)
	<-ex
}

type TrackingFunction func(inChan chan simplejson.Json, route string, port int)

func TrackingBlock(inTopic string, channel string, route string, port int, f TrackingFunction) {
	ex := make(chan bool)
	inChan := make(chan simplejson.Json)
	go nsqReader(inTopic, channel, inChan)
	go f(inChan, route, port)
	<-ex
}
