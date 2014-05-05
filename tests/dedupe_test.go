package tests

import (
	"log"
	"time"

	"github.com/nytlabs/streamtools/st/blocks"
	"github.com/nytlabs/streamtools/st/loghub"
	"github.com/nytlabs/streamtools/test_utils"
	. "launchpad.net/gocheck"
)

type DeDupeSuite struct{}

var deDupeSuite = Suite(&DeDupeSuite{})

func (s *DeDupeSuite) TestDeDupe(c *C) {
	loghub.Start()
	log.Println("testing dedupe")
	b, ch := test_utils.NewBlock("testing dedupe", "dedupe")

	emittedValues := make(map[string]bool)
	go blocks.BlockRoutine(b)
	outChan := make(chan *blocks.Msg)
	ch.AddChan <- &blocks.AddChanMsg{
		Route:   "out",
		Channel: outChan,
	}
	ruleMsg := map[string]interface{}{"Path": ".a"}
	rule := &blocks.Msg{Msg: ruleMsg, Route: "rule"}
	ch.InChan <- rule

	var sampleInput = map[string]interface{}{
		"a": "foobar",
	}

	time.AfterFunc(time.Duration(2)*time.Second, func() {
		postData := &blocks.Msg{Msg: sampleInput, Route: "in"}
		ch.InChan <- postData
	})

	time.AfterFunc(time.Duration(1)*time.Second, func() {
		postData := &blocks.Msg{Msg: sampleInput, Route: "in"}
		ch.InChan <- postData
	})

	time.AfterFunc(time.Duration(1)*time.Second, func() {
		postData := &blocks.Msg{Msg: map[string]interface{}{"a": "baz"}, Route: "in"}
		ch.InChan <- postData
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
		case messageI := <-outChan:
			message := messageI.Msg.(map[string]interface{})
			value := message["a"].(string)
			_, ok := emittedValues[value]
			if ok {
				c.Errorf("block emitted a dupe message", value)
			} else {
				emittedValues[value] = true
			}
		}
	}
}
