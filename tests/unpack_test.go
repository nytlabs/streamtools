package tests

import (
	"log"
	"time"

	"github.com/nytlabs/streamtools/st/blocks"
	"github.com/nytlabs/streamtools/st/loghub"
	"github.com/nytlabs/streamtools/test_utils"
	. "launchpad.net/gocheck"
)

type UnpackSuite struct{}

var unpackSuite = Suite(&UnpackSuite{})

func (s *UnpackSuite) TestUnpack(c *C) {
	loghub.Start()
	log.Println("testing unpack")
	b, ch := test_utils.NewBlock("testingunpack", "unpack")
	go blocks.BlockRoutine(b)
	outChan := make(chan *blocks.Msg)
	ch.AddChan <- &blocks.AddChanMsg{
		Route:   "out",
		Channel: outChan,
	}
	time.AfterFunc(time.Duration(5)*time.Second, func() {
		ch.QuitChan <- true
	})
	ruleMsg := map[string]interface{}{"Path": ".a"}
	rule := &blocks.Msg{Msg: ruleMsg, Route: "rule"}
	ch.InChan <- rule
	m := map[string]string{"b": "test"}
	arr := []interface{}{m}
	inMsg := map[string]interface{}{"a": arr}
	ch.InChan <- &blocks.Msg{Msg: inMsg, Route: "in"}
	for {
		select {
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
