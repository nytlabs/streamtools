package tests

import (
	"log"
	"reflect"
	"time"

	"github.com/nytlabs/streamtools/st/blocks"
	"github.com/nytlabs/streamtools/test_utils"
	. "launchpad.net/gocheck"
)

type ParseXMLSuite struct{}

var parseXMLSuite = Suite(&ParseXMLSuite{})

func (s *ParseXMLSuite) TestParseXML(c *C) {
	log.Println("testing ParseXML")
	b, ch := test_utils.NewBlock("testingParseXML", "parsexml")
	go blocks.BlockRoutine(b)
	outChan := make(chan *blocks.Msg)
	ch.AddChan <- &blocks.AddChanMsg{
		Route:   "out",
		Channel: outChan,
	}

	// where to find the xml in input
	ruleMsg := map[string]interface{}{"Path": ".data"}
	toRule := &blocks.Msg{Msg: ruleMsg, Route: "rule"}
	ch.InChan <- toRule

	queryOutChan := make(blocks.MsgChan)
	time.AfterFunc(time.Duration(1)*time.Second, func() {
		ch.QueryChan <- &blocks.QueryMsg{MsgChan: queryOutChan, Route: "rule"}
	})

	var xmldata = string(`
  <?xml version="1.0" encoding="utf-8"?>
  <OdfBody DocumentType="DT_GM" Date="20130131" Time="140807885" LogicalDate="20130131" Venue="ACV" Language="ENG" FeedFlag="P" DocumentCode="AS0ACV000" Version="3" Serial="1">
    <Competition Code="OG2014">
      <Config SDelay="60" />
    </Competition>
  </OdfBody>
  `)

	time.AfterFunc(time.Duration(2)*time.Second, func() {
		xmlMsg := map[string]interface{}{"data": xmldata}
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
			odfbody := message["OdfBody"].(map[string]interface{})
			competition := odfbody["Competition"].(map[string]interface{})
			c.Assert(odfbody["DocumentType"], Equals, "DT_GM")
			c.Assert(competition["Code"], Equals, "OG2014")
		}
	}
}
