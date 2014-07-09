package tests

import (
	"log"
	"time"

	"github.com/nytlabs/streamtools/st/blocks"
	"github.com/nytlabs/streamtools/test_utils"
	. "launchpad.net/gocheck"
)

type NSQSuite struct{}

var nsqSuite = Suite(&NSQSuite{})

func (s *NSQSuite) TestToNSQ(c *C) {
	log.Println("testing toNSQ")

	toB, toC := test_utils.NewBlock("testingToNSQ", "tonsq")
	go blocks.BlockRoutine(toB)

	ruleMsg := map[string]interface{}{"Topic": "librarytest", "NsqdTCPAddrs": "127.0.0.1:4150"}
	toRule := &blocks.Msg{Msg: ruleMsg, Route: "rule"}
	toC.InChan <- toRule

	toQueryChan := make(blocks.MsgChan)
	time.AfterFunc(time.Duration(1)*time.Second, func() {
		toC.QueryChan <- &blocks.QueryMsg{MsgChan: toQueryChan, Route: "rule"}
	})

	outChan := make(chan *blocks.Msg)
	toC.AddChan <- &blocks.AddChanMsg{Route: "1", Channel: outChan}

	time.AfterFunc(time.Duration(2)*time.Second, func() {
		nsqMsg := map[string]interface{}{"Foo": "Bar"}
		postData := &blocks.Msg{Msg: nsqMsg, Route: "in"}
		toC.InChan <- postData
	})

	time.AfterFunc(time.Duration(5)*time.Second, func() {
		toC.QuitChan <- true
	})

	for {
		select {
		case messageI := <-toQueryChan:
			c.Assert(messageI, DeepEquals, ruleMsg)

		case message := <-outChan:
			log.Println("printing message from outChan:", message)

		case err := <-toC.ErrChan:
			if err != nil {
				c.Errorf(err.Error())
			} else {
				return
			}
		}
	}
}

func (s *NSQSuite) TestFromNSQ(c *C) {
	log.Println("testing fromNSQ")

	fromB, fromC := test_utils.NewBlock("testingfromNSQ", "fromnsq")
	go blocks.BlockRoutine(fromB)

	outChan := make(chan *blocks.Msg)
	fromC.AddChan <- &blocks.AddChanMsg{Route: "1", Channel: outChan}

	var maxInFlight float64 = 100
	nsqSetup := map[string]interface{}{"ReadTopic": "librarytest", "LookupdAddr": "127.0.0.1:4161", "ReadChannel": "libtestchannel", "MaxInFlight": maxInFlight}
	fromRule := &blocks.Msg{Msg: nsqSetup, Route: "rule"}
	fromC.InChan <- fromRule

	fromQueryChan := make(blocks.MsgChan)
	time.AfterFunc(time.Duration(2)*time.Second, func() {
		fromC.QueryChan <- &blocks.QueryMsg{MsgChan: fromQueryChan, Route: "rule"}
	})

	time.AfterFunc(time.Duration(5)*time.Second, func() {
		fromC.QuitChan <- true
	})

	for {
		select {
		case messageI := <-fromQueryChan:
			c.Assert(messageI, DeepEquals, nsqSetup)

		case message := <-outChan:
			log.Println("printing message from outChan:", message)

		case err := <-fromC.ErrChan:
			if err != nil {
				c.Errorf(err.Error())
			} else {
				return
			}
		}
	}
}
