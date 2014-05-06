package tests

import (
	"log"
	"reflect"
	"testing"
	"time"

	"github.com/nytlabs/streamtools/st/blocks"
	"github.com/nytlabs/streamtools/test_utils"
	. "launchpad.net/gocheck"
)

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { TestingT(t) }

type CountSuite struct{}

var countSuite = Suite(&CountSuite{})

func (s *CountSuite) TestCount(c *C) {
	log.Println("testing Count")
	b, ch := test_utils.NewBlock("testingCount", "count")
	go blocks.BlockRoutine(b)

	ruleMsg := map[string]interface{}{"Window": "1s"}
	toRule := &blocks.Msg{Msg: ruleMsg, Route: "rule"}
	ch.InChan <- toRule

	inMsgMsg := map[string]interface{}{}
	inMsg := &blocks.Msg{Msg: inMsgMsg, Route: "in"}
	ch.InChan <- inMsg

	outChan := make(chan *blocks.Msg)
	ch.AddChan <- &blocks.AddChanMsg{Route: "1", Channel: outChan}

	queryOutChan := make(blocks.MsgChan)
	ch.QueryChan <- &blocks.QueryMsg{MsgChan: queryOutChan, Route: "rule"}

	countChan := make(blocks.MsgChan)
	ch.QueryChan <- &blocks.QueryMsg{MsgChan: countChan, Route: "count"}

	time.AfterFunc(time.Duration(5)*time.Second, func() {
		ch.QuitChan <- true
	})

	pollMsg := map[string]interface{}{}
	toPoll := &blocks.Msg{Msg: pollMsg, Route: "poll"}
	ch.InChan <- toPoll

	testOutput := map[string]interface{}{
		"Count": 1.0,
	}

	for {
		select {
		case messageI := <-queryOutChan:
			if !reflect.DeepEqual(messageI, ruleMsg) {
				c.Fail()
			}

		case messageI := <-countChan:
			if !reflect.DeepEqual(messageI, testOutput) {
				log.Println("count mismatch", messageI, testOutput)
				c.Fail()
			}

		case messageI := <-outChan:
			if !reflect.DeepEqual(messageI.Msg, testOutput) {
				log.Println("poll mismatch", messageI.Msg, testOutput)
				c.Fail()
			}

		case err := <-ch.ErrChan:
			if err != nil {
				c.Errorf(err.Error())
			} else {
				return
			}
		}
	}
}
