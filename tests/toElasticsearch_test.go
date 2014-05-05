package tests

import (
	"log"
	"reflect"
	"time"

	"github.com/nytlabs/streamtools/st/blocks"
	"github.com/nytlabs/streamtools/test_utils"
	. "launchpad.net/gocheck"
)

type ToElasticsearchSuite struct{}

var toElasticsearchSuite = Suite(&ToElasticsearchSuite{})

func (s *ToElasticsearchSuite) TestToElasticsearch(c *C) {
	log.Println("testing ToElasticsearch")
	b, ch := test_utils.NewBlock("testingToElasticsearch", "toelasticsearch")
	go blocks.BlockRoutine(b)
	outChan := make(chan *blocks.Msg)
	ch.AddChan <- &blocks.AddChanMsg{
		Route:   "out",
		Channel: outChan,
	}

	ruleMsg := map[string]interface{}{"Host": "localhost", "Port": "9200", "Index": "librarytest", "IndexType": "foobar"}
	rule := &blocks.Msg{Msg: ruleMsg, Route: "rule"}
	ch.InChan <- rule

	queryOutChan := make(blocks.MsgChan)
	time.AfterFunc(time.Duration(1)*time.Second, func() {
		ch.QueryChan <- &blocks.QueryMsg{MsgChan: queryOutChan, Route: "rule"}
	})

	time.AfterFunc(time.Duration(5)*time.Second, func() {
		ch.QuitChan <- true
	})
	for {
		select {
		case messageI := <-queryOutChan:
			if !reflect.DeepEqual(messageI, ruleMsg) {
				log.Println("rule mismatch:", messageI, ruleMsg)
				c.Fail()
			}

		case err := <-ch.ErrChan:
			if err != nil {
				c.Errorf(err.Error())
			} else {
				return
			}
		case <-outChan:
		}
	}
}
