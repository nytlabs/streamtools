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
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		fmt.Fprint(w, okResponse)
	}))

	defer ts.Close()

	headers := map[string]interface{}{"Content-Type": "application/json"}
	ruleMsg := map[string]interface{}{"Url": ts.URL, "UrlPath": "", "BodyPath": ".", "Method": "POST", "Headers": headers}
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
			messageHeaders := message["headers"].(http.Header)
			c.Assert(message["body"], NotNil)
			c.Assert(message["headers"], NotNil)
			c.Assert(messageHeaders.Get("Content-Type"), Equals, "application/json; charset=utf-8")
			c.Assert(message["status"], Equals, "200 OK")
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
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		fmt.Fprint(w, okResponse)
	}))

	defer ts.Close()

	headers := map[string]interface{}{"Content-Type": "application/json"}
	ruleMsg := map[string]interface{}{"Url": ts.URL, "UrlPath": "", "BodyPath": ".", "Method": "GET", "Headers": headers}
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
			messageHeaders := message["headers"].(http.Header)
			c.Assert(message["body"], NotNil)
			c.Assert(message["headers"], NotNil)
			c.Assert(messageHeaders.Get("Content-Type"), Equals, "application/json; charset=utf-8")
			c.Assert(message["status"], Equals, "200 OK")
		}
	}
}

func (s *WebRequestSuite) TestWebRequestGetXML(c *C) {
	log.Println("testing WebRequest: GET with XML")
	b, ch := test_utils.NewBlock("testingWebRequestGetXML", "webRequest")
	go blocks.BlockRoutine(b)
	outChan := make(chan *blocks.Msg)
	ch.AddChan <- &blocks.AddChanMsg{
		Route:   "out",
		Channel: outChan,
	}

	var xmldata = string(`
  <?xml version="1.0" encoding="utf-8"?>
  <OdfBody DocumentType="DT_GM" Date="20130131" Time="140807885" LogicalDate="20130131" Venue="ACV" Language="ENG" FeedFlag="P" DocumentCode="AS0ACV000" Version="3" Serial="1">
    <Competition Code="OG2014">
      <Config SDelay="60" />
    </Competition>
  </OdfBody>
  `)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/xml; charset=utf-8")
		fmt.Fprint(w, xmldata)
	}))

	defer ts.Close()

	headers := map[string]interface{}{"Content-Type": "application/xml"}
	ruleMsg := map[string]interface{}{"Url": ts.URL, "UrlPath": "", "BodyPath": ".", "Method": "GET", "Headers": headers}
	toRule := &blocks.Msg{Msg: ruleMsg, Route: "rule"}
	ch.InChan <- toRule

	queryOutChan := make(blocks.MsgChan)
	time.AfterFunc(time.Duration(1)*time.Second, func() {
		ch.QueryChan <- &blocks.QueryMsg{MsgChan: queryOutChan, Route: "rule"}
	})

	time.AfterFunc(time.Duration(2)*time.Second, func() {
		emptyMsg := make(map[string]interface{})
		postData := &blocks.Msg{Msg: emptyMsg, Route: "in"}
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
			messageHeaders := message["headers"].(http.Header)
			c.Assert(message["body"], NotNil)
			c.Assert(message["headers"], NotNil)
			c.Assert(messageHeaders.Get("Content-Type"), Equals, "text/xml; charset=utf-8")
			c.Assert(message["status"], Equals, "200 OK")
		}
	}
}

func (s *WebRequestSuite) TestWebRequestGetUrlPath(c *C) {
	log.Println("testing WebRequest: GET with UrlPath")
	b, ch := test_utils.NewBlock("testingWebRequestGetUrlPath", "webRequest")
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
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		fmt.Fprint(w, okResponse)
	}))

	defer ts.Close()

	headers := map[string]interface{}{"Content-Type": "application/json"}
	ruleMsg := map[string]interface{}{"Url": "", "UrlPath": ".url", "BodyPath": ".", "Method": "GET", "Headers": headers}
	toRule := &blocks.Msg{Msg: ruleMsg, Route: "rule"}
	ch.InChan <- toRule

	queryOutChan := make(blocks.MsgChan)
	time.AfterFunc(time.Duration(1)*time.Second, func() {
		ch.QueryChan <- &blocks.QueryMsg{MsgChan: queryOutChan, Route: "rule"}
	})

	time.AfterFunc(time.Duration(2)*time.Second, func() {
		urlPathMsg := map[string]interface{}{"url": ts.URL}
		postData := &blocks.Msg{Msg: urlPathMsg, Route: "in"}
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
			messageHeaders := message["headers"].(http.Header)
			c.Assert(message["body"], NotNil)
			c.Assert(message["headers"], NotNil)
			c.Assert(messageHeaders.Get("Content-Type"), Equals, "application/json; charset=utf-8")
			c.Assert(message["status"], Equals, "200 OK")
		}
	}
}

func (s *WebRequestSuite) TestWebRequestPostUrlPath(c *C) {
	log.Println("testing WebRequest: POST with UrlPath")
	b, ch := test_utils.NewBlock("testingWebRequestPostUrlPath", "webRequest")
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
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		fmt.Fprint(w, okResponse)
	}))

	defer ts.Close()

	headers := map[string]interface{}{"Content-Type": "application/json"}
	ruleMsg := map[string]interface{}{"Url": "", "UrlPath": ".url", "BodyPath": ".foo", "Method": "POST", "Headers": headers}
	toRule := &blocks.Msg{Msg: ruleMsg, Route: "rule"}
	ch.InChan <- toRule

	queryOutChan := make(blocks.MsgChan)
	time.AfterFunc(time.Duration(1)*time.Second, func() {
		ch.QueryChan <- &blocks.QueryMsg{MsgChan: queryOutChan, Route: "rule"}
	})

	time.AfterFunc(time.Duration(2)*time.Second, func() {
		urlPathMsg := map[string]interface{}{"url": ts.URL, "foo": "bar"}
		postData := &blocks.Msg{Msg: urlPathMsg, Route: "in"}
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
			messageHeaders := message["headers"].(http.Header)
			c.Assert(message["body"], NotNil)
			c.Assert(message["headers"], NotNil)
			c.Assert(messageHeaders.Get("Content-Type"), Equals, "application/json; charset=utf-8")
			c.Assert(message["status"], Equals, "200 OK")
		}
	}
}
