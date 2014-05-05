package tests

import (
	"log"
	"reflect"
	"time"

	"github.com/nytlabs/streamtools/st/blocks"
	"github.com/nytlabs/streamtools/st/loghub"
	"github.com/nytlabs/streamtools/test_utils"
	. "launchpad.net/gocheck"
)

type TickerSuite struct{}

var tickerSuite = Suite(&TickerSuite{})

func (s *TickerSuite) TestTicker(c *C) {
	loghub.Start()
	log.Println("testing Ticker")
	b, ch := test_utils.NewBlock("testingTicker", "ticker")
	go blocks.BlockRoutine(b)

	time.AfterFunc(time.Duration(5)*time.Second, func() {
		ch.QuitChan <- true
	})

	outChan := make(chan *blocks.Msg)
	ch.AddChan <- &blocks.AddChanMsg{Route: "1", Channel: outChan}

	queryOutChan := make(blocks.MsgChan)
	ch.QueryChan <- &blocks.QueryMsg{MsgChan: queryOutChan, Route: "rule"}

	ruleMsg := map[string]interface{}{"Interval": "1s"}
	toRule := &blocks.Msg{Msg: ruleMsg, Route: "rule"}
	ch.InChan <- toRule

	for {
		select {
		case err := <-ch.ErrChan:
			if err != nil {
				c.Errorf(err.Error())
			} else {
				return
			}
		case messageI := <-queryOutChan:
			if !reflect.DeepEqual(messageI, ruleMsg) {
				log.Println("Rule mismatch:", messageI, ruleMsg)
				c.Fail()
			}

		case messageI := <-outChan:
			message := messageI.Msg.(map[string]interface{})
			c.Assert(message["tick"], NotNil)
		}
	}
}
