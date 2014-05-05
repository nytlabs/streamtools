package tests

import (
	"log"
	"time"

	"github.com/nytlabs/streamtools/st/blocks"
	"github.com/nytlabs/streamtools/st/loghub"
	"github.com/nytlabs/streamtools/test_utils"
	. "launchpad.net/gocheck"
)

type SetSuite struct{}

var setSuite = Suite(&SetSuite{})

func (s *SetSuite) TestSet(c *C) {
	loghub.Start()
	log.Println("testing set")
	b, ch := test_utils.NewBlock("testing set", "set")
	go blocks.BlockRoutine(b)
	outChan := make(chan *blocks.Msg)
	ch.AddChan <- &blocks.AddChanMsg{
		Route:   "out",
		Channel: outChan,
	}
	ruleMsg := map[string]interface{}{"Path": ".a"}
	rule := &blocks.Msg{Msg: ruleMsg, Route: "rule"}
	ch.InChan <- rule
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
		case <-outChan:
		}
	}
}
