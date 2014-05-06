package tests

import (
	"log"
	"reflect"
	"time"

	"github.com/nytlabs/streamtools/st/blocks"
	"github.com/nytlabs/streamtools/test_utils"
	. "launchpad.net/gocheck"
)

type GetHTTPSuite struct{}

var getHTTPSuite = Suite(&GetHTTPSuite{})

func (s *GetHTTPSuite) TestGetHTTP(c *C) {
	log.Println("testing GetHTTP")
	b, ch := test_utils.NewBlock("testingGetHTTP", "gethttp")
	go blocks.BlockRoutine(b)
	outChan := make(chan *blocks.Msg)
	ch.AddChan <- &blocks.AddChanMsg{
		Route:   "out",
		Channel: outChan,
	}

	ruleMsg := map[string]interface{}{"Path": ".url"}
	toRule := &blocks.Msg{Msg: ruleMsg, Route: "rule"}
	ch.InChan <- toRule

	queryOutChan := make(blocks.MsgChan)
	time.AfterFunc(time.Duration(1)*time.Second, func() {
		ch.QueryChan <- &blocks.QueryMsg{MsgChan: queryOutChan, Route: "rule"}
	})

	time.AfterFunc(time.Duration(2)*time.Second, func() {
		nsqMsg := map[string]interface{}{"url": "https://raw.github.com/nytlabs/streamtools/master/examples/citibike.json"}
		postData := &blocks.Msg{Msg: nsqMsg, Route: "in"}
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
		case messageI := <-queryOutChan:
			if !reflect.DeepEqual(messageI, ruleMsg) {
				log.Println("Rule mismatch:", messageI, ruleMsg)
				c.Fail()
			}
		case msg := <-outChan:
			log.Println(msg)
		}
	}
}

func (s *GetHTTPSuite) TestGetHTTPXML(c *C) {
	log.Println("testing GetHTTP with XML")
	b, ch := test_utils.NewBlock("testingGetHTTPXML", "gethttp")
	go blocks.BlockRoutine(b)
	outChan := make(chan *blocks.Msg)
	ch.AddChan <- &blocks.AddChanMsg{
		Route:   "out",
		Channel: outChan,
	}

	ruleMsg := map[string]interface{}{"Path": ".url"}
	toRule := &blocks.Msg{Msg: ruleMsg, Route: "rule"}
	ch.InChan <- toRule

	queryOutChan := make(blocks.MsgChan)
	time.AfterFunc(time.Duration(1)*time.Second, func() {
		ch.QueryChan <- &blocks.QueryMsg{MsgChan: queryOutChan, Route: "rule"}
	})

	time.AfterFunc(time.Duration(2)*time.Second, func() {
		xmlMsg := map[string]interface{}{"url": "https://raw.github.com/nytlabs/streamtools/master/examples/odf.xml"}
		postData := &blocks.Msg{Msg: xmlMsg, Route: "in"}
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
		case messageI := <-queryOutChan:
			if !reflect.DeepEqual(messageI, ruleMsg) {
				log.Println("Rule mismatch:", messageI, ruleMsg)
				c.Fail()
			}
		case messageI := <-outChan:
			message := messageI.Msg.(map[string]interface{})
			messageData := message["data"].(string)
			var xmldata = string(`<?xml version="1.0" encoding="utf-8"?>
<OdfBody DocumentType="DT_GM" Date="20130131" Time="140807885" LogicalDate="20130131" Venue="ACV" Language="ENG" FeedFlag="P" DocumentCode="AS0ACV000" Version="3" Serial="1">
	<Competition Code="OG2014">
		<Config SDelay="60" />
	</Competition>
</OdfBody>
`)
			c.Assert(messageData, Equals, xmldata)
		}
	}
}
