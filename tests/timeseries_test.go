package tests

import (
	"log"
	"time"

	"github.com/nytlabs/streamtools/st/blocks"
	"github.com/nytlabs/streamtools/test_utils"
	. "launchpad.net/gocheck"
)

type TimeseriesSuite struct{}

var timeseriesSuite = Suite(&TimeseriesSuite{})

func (s *TimeseriesSuite) TestTimeseries(c *C) {
	log.Println("testing Timeseries")
	b, ch := test_utils.NewBlock("testingTimeseries", "timeseries")
	go blocks.BlockRoutine(b)
	outChan := make(chan *blocks.Msg)
	ch.AddChan <- &blocks.AddChanMsg{
		Route:   "out",
		Channel: outChan,
	}
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
			log.Println("out")
		}
	}
}
