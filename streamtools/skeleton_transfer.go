package streamtools

import (
	"github.com/bitly/go-simplejson"
	"log"
)

func SkeletonTransfer(inChan chan *simplejson.Json, outChan chan *simplejson.Json, RuleChan chan *simplejson.Json) {

	// get rules
	rules := <-RuleChan
	log.Println("using rules", rules)

	for {
		select {
		case rules := <-RuleChan:
			// get new rules
			log.Println("using rules", rules)

		case msg := <-inChan:
			// process inbound message
			outMsg := msg

			// emit outbound message
			log.Println("emitting message", msg)
			outChan <- outMsg
		}
	}
}
