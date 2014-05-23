package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"reflect"
	"time"

	"github.com/nytlabs/streamtools/st/blocks"
	"github.com/nytlabs/streamtools/test_utils"
	. "launchpad.net/gocheck"
)

type WebRequestSuite struct{}

var webRequestSuite = Suite(&WebRequestSuite{})

func (s *WebRequestSuite) TestWebRequestPost(c *C) {
	log.Println("testing WebRequest: POST")
	b, ch := test_utils.NewBlock("testingWebRequestPost", "webRequest")
	go blocks.BlockRoutine(b)
	outChan := make(chan *blocks.Msg)
	ch.AddChan <- &blocks.AddChanMsg{
		Route:   "out",
		Channel: outChan,
	}

	var okResponse interface{}
	statusOk := bytes.NewBufferString(`{"Status": "OK"}`)
	err := json.Unmarshal(statusOk.Bytes(), &okResponse)
	if err != nil {
		log.Println("unable to unmarshal json")
		c.Fail()
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, okResponse)
	}))

	defer ts.Close()

	headers := map[string]interface{}{"Content-Type": "application/json"}
	ruleMsg := map[string]interface{}{"Url": ts.URL, "Method": "POST", "Headers": headers}
	toRule := &blocks.Msg{Msg: ruleMsg, Route: "rule"}
	ch.InChan <- toRule

	queryOutChan := make(blocks.MsgChan)
	time.AfterFunc(time.Duration(1)*time.Second, func() {
		ch.QueryChan <- &blocks.QueryMsg{MsgChan: queryOutChan, Route: "rule"}
	})

	time.AfterFunc(time.Duration(2)*time.Second, func() {
		foobarMsg := map[string]interface{}{"foo": "bar"}
		postData := &blocks.Msg{Msg: foobarMsg, Route: "in"}
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
			c.Assert(message["Response"], NotNil)
		}
	}
}

func (s *WebRequestSuite) TestWebRequestGet(c *C) {
	log.Println("testing WebRequest: GET")
	b, ch := test_utils.NewBlock("testingWebRequestGet", "webRequest")
	go blocks.BlockRoutine(b)
	outChan := make(chan *blocks.Msg)
	ch.AddChan <- &blocks.AddChanMsg{
		Route:   "out",
		Channel: outChan,
	}

	var okResponse interface{}
	statusOk := bytes.NewBufferString(`{"Status": "OK"}`)
	err := json.Unmarshal(statusOk.Bytes(), &okResponse)
	if err != nil {
		log.Println("unable to unmarshal json")
		c.Fail()
	}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, okResponse)
	}))

	defer ts.Close()

	headers := map[string]interface{}{"Content-Type": "application/json"}
	ruleMsg := map[string]interface{}{"Url": ts.URL, "Method": "GET", "Headers": headers}
	toRule := &blocks.Msg{Msg: ruleMsg, Route: "rule"}
	ch.InChan <- toRule

	queryOutChan := make(blocks.MsgChan)
	time.AfterFunc(time.Duration(1)*time.Second, func() {
		ch.QueryChan <- &blocks.QueryMsg{MsgChan: queryOutChan, Route: "rule"}
	})

	time.AfterFunc(time.Duration(2)*time.Second, func() {
		foobarMsg := map[string]interface{}{"foo": "bar"}
		postData := &blocks.Msg{Msg: foobarMsg, Route: "in"}
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
