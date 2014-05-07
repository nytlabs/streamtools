package tests

import (
	"log"
	"time"

	"github.com/nytlabs/streamtools/st/blocks"
	"github.com/nytlabs/streamtools/st/loghub"
	"github.com/nytlabs/streamtools/test_utils"
	. "launchpad.net/gocheck"
)

type CacheSuite struct{}

var cacheSuite = Suite(&CacheSuite{})

func (s *CacheSuite) TestCache(c *C) {
	loghub.Start()
	log.Println("testing cache")
	b, ch := test_utils.NewBlock("testing cache", "cache")

	go blocks.BlockRoutine(b)
	outChan := make(chan *blocks.Msg)
	ch.AddChan <- &blocks.AddChanMsg{
		Route:   "out",
		Channel: outChan,
	}
	ruleMsg := map[string]interface{}{"KeyPath": ".name", "ValuePath": ".count", "TimeToLive": "1m"}
	rule := &blocks.Msg{Msg: ruleMsg, Route: "rule"}
	ch.InChan <- rule

	// Add some data to the cache
	time.AfterFunc(time.Duration(1)*time.Second, func() {
		ch.InChan <- &blocks.Msg{Msg: map[string]interface{}{"count": "100", "name": "The New York Times"}, Route: "in"}
		ch.InChan <- &blocks.Msg{Msg: map[string]interface{}{"count": "4", "name": "The New York Times"}, Route: "in"}
		ch.InChan <- &blocks.Msg{Msg: map[string]interface{}{"count": "50", "name": "Hacks/Hackers"}, Route: "in"}
	})

	// Query for keys
	keysChan := make(blocks.MsgChan)
	time.AfterFunc(time.Duration(2)*time.Second, func() {
		ch.QueryChan <- &blocks.QueryMsg{MsgChan: keysChan, Route: "keys"}
	})

	// And values
	valuesChan := make(blocks.MsgChan)
	time.AfterFunc(time.Duration(2)*time.Second, func() {
		ch.QueryChan <- &blocks.QueryMsg{MsgChan: valuesChan, Route: "values"}
	})

	// And the entire cache contents
	dumpChan := make(blocks.MsgChan)
	time.AfterFunc(time.Duration(2)*time.Second, func() {
		ch.QueryChan <- &blocks.QueryMsg{MsgChan: dumpChan, Route: "dump"}
	})

	time.AfterFunc(time.Duration(5)*time.Second, func() {
		ch.QuitChan <- true
	})
	for {
		select {
		case err := <-ch.ErrChan:
			if err != nil {
				c.Errorf(err.Error())
			} else {
				return
			}

		case messageI := <-keysChan:
			message := messageI.(map[string]interface{})
			keys := message["keys"].([]string)
			c.Assert(keys, DeepEquals, []string{"The New York Times", "Hacks/Hackers"})

		case messageI := <-valuesChan:
			message := messageI.(map[string]interface{})
			values := message["values"].([]interface{})
			c.Assert(values, HasLen, 2)

		case messageI := <-dumpChan:
			message := messageI.(map[string]interface{})
			c.Assert(message["dump"], HasLen, 2)

		case messageI := <-outChan:
			message := messageI.Msg.(map[string]interface{})
			log.Println(message)
		}
	}
}
