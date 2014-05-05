package tests

import (
	"log"
	"reflect"
	"time"

	"github.com/nytlabs/streamtools/st/blocks"
	"github.com/nytlabs/streamtools/test_utils"
	. "launchpad.net/gocheck"
)

type MapSuite struct{}

var mapSuite = Suite(&MapSuite{})

func (s *MapSuite) TestMap(c *C) {
	log.Println("testing Map")
	b, ch := test_utils.NewBlock("testingMap", "map")
	go blocks.BlockRoutine(b)

	outChan := make(chan *blocks.Msg)
	ch.AddChan <- &blocks.AddChanMsg{Route: "1", Channel: outChan}

	//outChan := make(chan *blocks.Msg)
	//ch.AddChan <- &blocks.AddChanMsg{
	//	Route:   "out",
	//	Channel: outChan,
	//}

	mapMsg := map[string]interface{}{"MegaBar": ".bar"}
	ruleMsg := map[string]interface{}{"Map": mapMsg, "Additive": false}

	// send rule
	toRule := &blocks.Msg{Msg: ruleMsg, Route: "rule"}
	ch.InChan <- toRule

	// send rule query in a sec
	queryOutChan := make(blocks.MsgChan)
	time.AfterFunc(time.Duration(1)*time.Second, func() {
		ch.QueryChan <- &blocks.QueryMsg{MsgChan: queryOutChan, Route: "rule"}
	})

	// quit after 5
	time.AfterFunc(time.Duration(5)*time.Second, func() {
		ch.QuitChan <- true
	})

	// send test input
	inputMsg := map[string]interface{}{"bar": "something", "foo": "another thing"}
	inputBlock := &blocks.Msg{Msg: inputMsg, Route: "in"}
	time.AfterFunc(time.Duration(2)*time.Second, func() {
		ch.InChan <- inputBlock
	})

	for {
		select {
		case messageI := <-queryOutChan:
			message := messageI.(map[string]interface{})
			if !reflect.DeepEqual(message["Map"], ruleMsg["Map"]) {
				log.Println(messageI)
				log.Println("was expecting", ruleMsg["Map"], "but got", message["Map"])
				c.Fail()
			}
		case err := <-ch.ErrChan:
			if err != nil {
				c.Errorf(err.Error())
			} else {
				return
			}
		case messageI := <-outChan:
			message := messageI.Msg.(map[string]interface{})
			log.Println(message)
			c.Assert(message["MegaBar"], Equals, "something")
			c.Assert(message["foo"], IsNil)
		}
	}
}
