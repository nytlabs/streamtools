package tests

import (
	"log"
	"os"
	"reflect"
	"time"

	"github.com/nytlabs/streamtools/st/blocks"
	"github.com/nytlabs/streamtools/st/loghub"
	"github.com/nytlabs/streamtools/test_utils"
	. "launchpad.net/gocheck"
)

type ToFileSuite struct{}

var toFileSuite = Suite(&ToFileSuite{})

func (s *ToFileSuite) TestToFile(c *C) {
	loghub.Start()
	log.Println("testing toFile")
	b, ch := test_utils.NewBlock("testingToFile", "tofile")
	go blocks.BlockRoutine(b)

	ruleMsg := map[string]interface{}{"Filename": "foobar.log"}
	toRule := &blocks.Msg{Msg: ruleMsg, Route: "rule"}
	ch.InChan <- toRule

	outChan := make(chan *blocks.Msg)
	ch.AddChan <- &blocks.AddChanMsg{Route: "1", Channel: outChan}

	queryOutChan := make(blocks.MsgChan)
	ch.QueryChan <- &blocks.QueryMsg{MsgChan: queryOutChan, Route: "rule"}

	time.AfterFunc(time.Duration(5)*time.Second, func() {
		err := os.Remove("foobar.log")
		if err != nil {
			c.Errorf(err.Error())
		}
		ch.QuitChan <- true
	})

	for {
		select {
		case messageI := <-queryOutChan:
			if !reflect.DeepEqual(messageI, ruleMsg) {
				c.Fail()
			}

		case message := <-outChan:
			log.Println(message)

		case err := <-ch.ErrChan:
			if err != nil {
				c.Errorf(err.Error())
			} else {
				return
			}
		}
	}
}
