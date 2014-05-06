package tests

import (
	"io/ioutil"
	"log"
	"syscall"
	"time"

	"github.com/nytlabs/streamtools/st/blocks"
	"github.com/nytlabs/streamtools/test_utils"
	. "launchpad.net/gocheck"
)

type FromFileSuite struct{}

var fromFileSuite = Suite(&FromFileSuite{})

func (s *FromFileSuite) TestFromFile(c *C) {
	log.Println("testing FromFile")
	b, ch := test_utils.NewBlock("testingFile", "fromfile")
	go blocks.BlockRoutine(b)
	outChan := make(chan *blocks.Msg)
	ch.AddChan <- &blocks.AddChanMsg{
		Route:   "out",
		Channel: outChan,
	}

	f, err := ioutil.TempFile("", "streamtools_test_from_file.log")
	if err != nil {
		c.Errorf(err.Error())
	}
	defer syscall.Unlink(f.Name())
	var fromfilestring = string(`{"Name": "Jacqui Maher", "Location": "Brooklyn, NY", "Dog": "Conor S. Dogberst" }
{"Name": "Nik Hanselmann", "Location": "New York, NY", "Dog": "None:(" }
{"Name": "Mike Dewar", "Location": "Brooklyn, NY", "Dog": "Percy ? Dewar" }`)
	ioutil.WriteFile(f.Name(), []byte(fromfilestring), 0644)

	ruleMsg := map[string]interface{}{"Filename": f.Name()}
	toRule := &blocks.Msg{Msg: ruleMsg, Route: "rule"}
	ch.InChan <- toRule

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
