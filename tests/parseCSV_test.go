package tests

import (
	"log"
	"reflect"
	"time"

	"github.com/nytlabs/streamtools/st/blocks"
	"github.com/nytlabs/streamtools/test_utils"
	. "launchpad.net/gocheck"
)

type ParseCSVSuite struct{}

var parseCSVSuite = Suite(&ParseCSVSuite{})

func (s *ParseCSVSuite) TestParseCSV(c *C) {
	log.Println("testing ParseCSV")
	b, ch := test_utils.NewBlock("testingParseCSV", "parsecsv")
	go blocks.BlockRoutine(b)
	outChan := make(chan *blocks.Msg)
	ch.AddChan <- &blocks.AddChanMsg{
		Route:   "out",
		Channel: outChan,
	}

	// where to find the xml in input
	headers := []string{"name", "email", "phone"}
	ruleMsg := map[string]interface{}{"Path": ".data", "Headers": headers}
	toRule := &blocks.Msg{Msg: ruleMsg, Route: "rule"}
	ch.InChan <- toRule

	queryOutChan := make(blocks.MsgChan)
	time.AfterFunc(time.Duration(1)*time.Second, func() {
		ch.QueryChan <- &blocks.QueryMsg{MsgChan: queryOutChan, Route: "rule"}
	})

	var csvInput = `
	Jacqui Maher, jacqui@example.com, 555-5550
	Mike Dewar, mike@example.com, 555-5551
	Nik Hanselmann, nik@example.com, 555-5552
	Jane Friedoff, jane@example.com, 555-5553, Extra
	`

	time.AfterFunc(time.Duration(1)*time.Second, func() {
		csvMsg := map[string]interface{}{"data": csvInput}
		postData := &blocks.Msg{Msg: csvMsg, Route: "in"}
		ch.InChan <- postData
	})

	time.AfterFunc(time.Duration(1)*time.Second, func() {
		ch.InChan <- &blocks.Msg{Msg: map[string]interface{}{}, Route: "poll"}
	})
	time.AfterFunc(time.Duration(1)*time.Second, func() {
		ch.InChan <- &blocks.Msg{Msg: map[string]interface{}{}, Route: "poll"}
	})
	time.AfterFunc(time.Duration(1)*time.Second, func() {
		ch.InChan <- &blocks.Msg{Msg: map[string]interface{}{}, Route: "poll"}
	})
	time.AfterFunc(time.Duration(1)*time.Second, func() {
		ch.InChan <- &blocks.Msg{Msg: map[string]interface{}{}, Route: "poll"}
	})

	time.AfterFunc(time.Duration(6)*time.Second, func() {
		ch.QuitChan <- true
	})

	var expectedNames = []string{"Jacqui Maher", "Nik Hanselmann", "Mike Dewar", "Jane Friedoff"}

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

			log.Println(message)
			nameReceived := message["name"].(string)
			if !test_utils.StringInSlice(expectedNames, nameReceived) {
				log.Println("failed finding", nameReceived, "in expected names list")
				c.Fail()
			}
		}
	}
}
