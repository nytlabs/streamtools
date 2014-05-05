package tests

import (
	"log"
	"time"

	"github.com/nytlabs/streamtools/st/blocks"
	"github.com/nytlabs/streamtools/test_utils"
	. "launchpad.net/gocheck"
)

type FromPostSuite struct{}

var fromPostSuite = Suite(&FromPostSuite{})

func (s *FromPostSuite) TestFromPost(c *C) {
	log.Println("testing FromPost")
	b, ch := test_utils.NewBlock("testingPost", "frompost")
	go blocks.BlockRoutine(b)
	outChan := make(chan *blocks.Msg)
	ch.AddChan <- &blocks.AddChanMsg{
		Route:   "out",
		Channel: outChan,
	}

	inputMsg := map[string]interface{}{"Foo": "BAR"}
	inputBlock := &blocks.Msg{Msg: inputMsg, Route: "in"}
	ch.InChan <- inputBlock

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

func (s *FromPostSuite) TestFromPostXML(c *C) {
	log.Println("testing fromPost with XML")
	b, ch := test_utils.NewBlock("testingFromPostXML", "frompost")
	go blocks.BlockRoutine(b)
	outChan := make(chan *blocks.Msg)
	ch.AddChan <- &blocks.AddChanMsg{
		Route:   "out",
		Channel: outChan,
	}

	var xmlstring = string(`
  <?xml version="1.0" encoding="utf-8"?>
  <OdfBody DocumentType="DT_GM" Date="20130131" Time="140807885" LogicalDate="20130131" Venue="ACV" Language="ENG" FeedFlag="P" DocumentCode="AS0ACV000" Version="3" Serial="1">
    <Competition Code="OG2014">
      <Config SDelay="60" />
    </Competition>
  </OdfBody>
  `)

	var xmldata = map[string]interface{}{
		"data": xmlstring,
	}
	time.AfterFunc(time.Duration(2)*time.Second, func() {
		postData := &blocks.Msg{Msg: xmldata, Route: "in"}
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
			messageXML := message["data"].(string)
			c.Assert(messageXML, Equals, xmlstring)
		}
	}
}
