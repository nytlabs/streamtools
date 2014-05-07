package tests

import (
	"fmt"
	"log"
	"reflect"
	"time"

	"github.com/nytlabs/streamtools/st/blocks"
	"github.com/nytlabs/streamtools/test_utils"
	. "launchpad.net/gocheck"
)

type FromHTTPStreamSuite struct{}

var fromHTTPStreamSuite = Suite(&FromHTTPStreamSuite{})

func (s *FromHTTPStreamSuite) TestFromHTTPStreamXML(c *C) {
	log.Println("testing FromHTTPStream with XML")
	b, ch := test_utils.NewBlock("testingFromHTTPStreamXML", "fromhttpstream")
	go blocks.BlockRoutine(b)
	outChan := make(chan *blocks.Msg)
	ch.AddChan <- &blocks.AddChanMsg{
		Route:   "out",
		Channel: outChan,
	}

	ruleMsg := map[string]interface{}{"Endpoint": "https://raw.github.com/nytlabs/streamtools/master/examples/odf.xml"}
	toRule := &blocks.Msg{Msg: ruleMsg, Route: "rule"}
	ch.InChan <- toRule

	queryOutChan := make(blocks.MsgChan)
	time.AfterFunc(time.Duration(1)*time.Second, func() {
		ch.QueryChan <- &blocks.QueryMsg{MsgChan: queryOutChan, Route: "rule"}
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

		case messageI := <-queryOutChan:
			message := messageI.(map[string]interface{})
			if !reflect.DeepEqual(message["Endpoint"], ruleMsg["Endpoint"]) {
				log.Println("Rule mismatch:", message["Endpoint"], ruleMsg["Endpoint"])
				c.Fail()
			}

		case messageI := <-outChan:
			message := messageI.Msg.(map[string]interface{})
			fmt.Printf("%s", message["data"])
		}
	}
}

func (s *FromHTTPStreamSuite) TestFromHTTPStream(c *C) {
	log.Println("testing FromHTTPStream")
	b, ch := test_utils.NewBlock("testingFromHTTPStream", "fromhttpstream")
	go blocks.BlockRoutine(b)
	outChan := make(chan *blocks.Msg)
	ch.AddChan <- &blocks.AddChanMsg{
		Route:   "out",
		Channel: outChan,
	}

	ruleMsg := map[string]interface{}{"Endpoint": "http://www.nytimes.com"}
	toRule := &blocks.Msg{Msg: ruleMsg, Route: "rule"}
	ch.InChan <- toRule

	queryOutChan := make(blocks.MsgChan)
	time.AfterFunc(time.Duration(1)*time.Second, func() {
		ch.QueryChan <- &blocks.QueryMsg{MsgChan: queryOutChan, Route: "rule"}
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

		case messageI := <-queryOutChan:
			message := messageI.(map[string]interface{})
			if !reflect.DeepEqual(message["Endpoint"], ruleMsg["Endpoint"]) {
				log.Println("Rule mismatch:", message["Endpoint"], ruleMsg["Endpoint"])
				c.Fail()
			}

		case <-outChan:
		}
	}
}
