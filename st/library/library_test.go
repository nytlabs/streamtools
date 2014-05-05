package library

import (
	"fmt"
	"github.com/nytlabs/streamtools/st/blocks" // blocks
	"github.com/nytlabs/streamtools/st/loghub"
	"io/ioutil"
	. "launchpad.net/gocheck"
	"log"
	"net/http"
	"net/http/httptest"
	"reflect"
	"syscall"
	"testing"
	"time"
)

func newBlock(id, kind string) (blocks.BlockInterface, blocks.BlockChans) {

	chans := blocks.BlockChans{
		InChan:         make(chan *blocks.Msg),
		QueryChan:      make(chan *blocks.QueryMsg),
		QueryParamChan: make(chan *blocks.QueryParamMsg),
		AddChan:        make(chan *blocks.AddChanMsg),
		DelChan:        make(chan *blocks.Msg),
		ErrChan:        make(chan error),
		QuitChan:       make(chan bool),
	}

	// actual block
	newblock, ok := Blocks[kind]
	if !ok {
		log.Println("block", kind, "not found!")
	}
	b := newblock()
	b.Build(chans)

	return b, chans

}

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { TestingT(t) }

type StreamSuite struct{}

// this is run once before the entire test SUITE
func (s *StreamSuite) SetUpSuite(c *C) {
	loghub.Start()
}

// this would be run once before EACH of the tests
// func (s *StreamSuite) SetUpTest(c *C) {
//   // do something
// }

var _ = Suite(&StreamSuite{})

func (s *StreamSuite) TestToFromNSQ(c *C) {
	log.Println("testing toNSQ")

	toB, toC := newBlock("testingToNSQ", "tonsq")
	go blocks.BlockRoutine(toB)

	ruleMsg := map[string]interface{}{"Topic": "librarytest", "NsqdTCPAddrs": "127.0.0.1:4150"}
	toRule := &blocks.Msg{Msg: ruleMsg, Route: "rule"}
	toC.InChan <- toRule

	toQueryChan := make(blocks.MsgChan)
	time.AfterFunc(time.Duration(1)*time.Second, func() {
		toC.QueryChan <- &blocks.QueryMsg{MsgChan: toQueryChan, Route: "rule"}
	})

	time.AfterFunc(time.Duration(2)*time.Second, func() {
		nsqMsg := map[string]interface{}{"Foo": "Bar"}
		postData := &blocks.Msg{Msg: nsqMsg, Route: "in"}
		toC.InChan <- postData
	})

	time.AfterFunc(time.Duration(5)*time.Second, func() {
		toC.QuitChan <- true
	})

	log.Println("testing fromNSQ")

	fromB, fromC := newBlock("testingfromNSQ", "fromnsq")
	go blocks.BlockRoutine(fromB)

	outChan := make(chan *blocks.Msg)
	fromC.AddChan <- &blocks.AddChanMsg{Route: "1", Channel: outChan}

	var maxInFlight float64 = 100
	nsqSetup := map[string]interface{}{"ReadTopic": "librarytest", "LookupdAddr": "127.0.0.1:4161", "ReadChannel": "libtestchannel", "MaxInFlight": maxInFlight}
	fromRule := &blocks.Msg{Msg: nsqSetup, Route: "rule"}
	fromC.InChan <- fromRule

	fromQueryChan := make(blocks.MsgChan)
	time.AfterFunc(time.Duration(2)*time.Second, func() {
		fromC.QueryChan <- &blocks.QueryMsg{MsgChan: fromQueryChan, Route: "rule"}
	})

	time.AfterFunc(time.Duration(5)*time.Second, func() {
		fromC.QuitChan <- true
	})

	for {
		select {
		case messageI := <-fromQueryChan:
			c.Assert(messageI, DeepEquals, nsqSetup)

		case messageI := <-toQueryChan:
			c.Assert(messageI, DeepEquals, ruleMsg)

		case message := <-outChan:
			log.Println("printing message from outChan:", message)

		case err := <-toC.ErrChan:
			if err != nil {
				c.Errorf(err.Error())
			} else {
				return
			}
		case err := <-fromC.ErrChan:
			if err != nil {
				c.Errorf(err.Error())
			} else {
				return
			}
		}
	}
}

func (s *StreamSuite) TestCount(c *C) {
	log.Println("testing Count")
	b, ch := newBlock("testingCount", "count")
	go blocks.BlockRoutine(b)

	ruleMsg := map[string]interface{}{"Window": "1s"}
	toRule := &blocks.Msg{Msg: ruleMsg, Route: "rule"}
	ch.InChan <- toRule

	inMsgMsg := map[string]interface{}{}
	inMsg := &blocks.Msg{Msg: inMsgMsg, Route: "in"}
	ch.InChan <- inMsg

	outChan := make(chan *blocks.Msg)
	ch.AddChan <- &blocks.AddChanMsg{Route: "1", Channel: outChan}

	queryOutChan := make(blocks.MsgChan)
	ch.QueryChan <- &blocks.QueryMsg{MsgChan: queryOutChan, Route: "rule"}

	countChan := make(blocks.MsgChan)
	ch.QueryChan <- &blocks.QueryMsg{MsgChan: countChan, Route: "count"}

	time.AfterFunc(time.Duration(5)*time.Second, func() {
		ch.QuitChan <- true
	})

	pollMsg := map[string]interface{}{}
	toPoll := &blocks.Msg{Msg: pollMsg, Route: "poll"}
	ch.InChan <- toPoll

	testOutput := map[string]interface{}{
		"Count": 1.0,
	}

	for {
		select {
		case messageI := <-queryOutChan:
			if !reflect.DeepEqual(messageI, ruleMsg) {
				c.Fail()
			}

		case messageI := <-countChan:
			if !reflect.DeepEqual(messageI, testOutput) {
				log.Println("count mismatch", messageI, testOutput)
				c.Fail()
			}

		case messageI := <-outChan:
			if !reflect.DeepEqual(messageI.Msg, testOutput) {
				log.Println("poll mismatch", messageI.Msg, testOutput)
				c.Fail()
			}

		case err := <-ch.ErrChan:
			if err != nil {
				c.Errorf(err.Error())
			} else {
				return
			}
		}
	}
}

func (s *StreamSuite) TestToFile(c *C) {
	log.Println("testing toFile")
	b, ch := newBlock("testingToFile", "tofile")
	go blocks.BlockRoutine(b)

	ruleMsg := map[string]interface{}{"Filename": "foobar.log"}
	toRule := &blocks.Msg{Msg: ruleMsg, Route: "rule"}
	ch.InChan <- toRule

	outChan := make(chan *blocks.Msg)
	ch.AddChan <- &blocks.AddChanMsg{Route: "1", Channel: outChan}

	queryOutChan := make(blocks.MsgChan)
	ch.QueryChan <- &blocks.QueryMsg{MsgChan: queryOutChan, Route: "rule"}

	time.AfterFunc(time.Duration(5)*time.Second, func() {
		ch.QuitChan <- true
	})

	for {
		select {
		case messageI := <-queryOutChan:
			if !reflect.DeepEqual(messageI, ruleMsg) {
				c.Fail()
			}

		case message := <-outChan:
			log.Println(message)

		case err := <-ch.ErrChan:
			if err != nil {
				c.Errorf(err.Error())
			} else {
				return
			}
		}
	}
}

func (s *StreamSuite) TestFromSQS(c *C) {
	log.Println("testing FromSQS")

	sampleResponse := string(`
<CreateQueueResponse
  xmlns="http://sqs.us-east-1.amazonaws.com/doc/2012-11-05/"
  xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
  xsi:type="CreateQueueResponse">
     <CreateQueueResult>
        <QueueUrl>
        http://sqs.us-east-1.amazonaws.com/770098461991/queue2
        </QueueUrl>
     </CreateQueueResult>
     <ResponseMetadata>
        <RequestId>cb919c0a-9bce-4afe-9b48-9bdf2412bb67</RequestId>
     </ResponseMetadata>
</CreateQueueResponse>
  `)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, sampleResponse)
	}))
	defer ts.Close()

	b, ch := newBlock("testingFromSQS", "fromsqs")
	go blocks.BlockRoutine(b)

	ruleMsg := map[string]interface{}{"SQSEndpoint": ts.URL, "AccessKey": "123access", "AccessSecret": "123secret", "APIVersion": "2012-11-05", "SignatureVersion": "4", "WaitTimeSeconds": "0", "MaxNumberOfMessages": "10"}
	toRule := &blocks.Msg{Msg: ruleMsg, Route: "rule"}
	ch.InChan <- toRule

	outChan := make(chan *blocks.Msg)
	ch.AddChan <- &blocks.AddChanMsg{Route: "1", Channel: outChan}

	queryOutChan := make(blocks.MsgChan)
	ch.QueryChan <- &blocks.QueryMsg{MsgChan: queryOutChan, Route: "rule"}

	time.AfterFunc(time.Duration(5)*time.Second, func() {
		ch.QuitChan <- true
	})

	for {
		select {
		case messageI := <-queryOutChan:
			if !reflect.DeepEqual(messageI, ruleMsg) {
				log.Println("Rule mismatch:", messageI, ruleMsg)
				c.Fail()
			}

		case message := <-outChan:
			log.Println(message)

		case err := <-ch.ErrChan:
			if err != nil {
				c.Errorf(err.Error())
			} else {
				return
			}
		}
	}
}

func (s *StreamSuite) TestSync(c *C) {
	log.Println("testing Sync")
	b, ch := newBlock("testingSync", "sync")
	go blocks.BlockRoutine(b)
	time.AfterFunc(time.Duration(5)*time.Second, func() {
		ch.QuitChan <- true
	})
	err := <-ch.ErrChan
	if err != nil {
		c.Errorf(err.Error())
	}
}

func (s *StreamSuite) TestTicker(c *C) {
	log.Println("testing Ticker")
	b, ch := newBlock("testingTicker", "ticker")
	go blocks.BlockRoutine(b)

	time.AfterFunc(time.Duration(5)*time.Second, func() {
		ch.QuitChan <- true
	})

	outChan := make(chan *blocks.Msg)
	ch.AddChan <- &blocks.AddChanMsg{Route: "1", Channel: outChan}

	queryOutChan := make(blocks.MsgChan)
	ch.QueryChan <- &blocks.QueryMsg{MsgChan: queryOutChan, Route: "rule"}

	ruleMsg := map[string]interface{}{"Interval": "1s"}
	toRule := &blocks.Msg{Msg: ruleMsg, Route: "rule"}
	ch.InChan <- toRule

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
			c.Assert(message["tick"], NotNil)
		}
	}
}

func (s *StreamSuite) TestFilter(c *C) {
	log.Println("testing Filter")
	b, ch := newBlock("testingFilter", "filter")
	go blocks.BlockRoutine(b)

	ruleMsg := map[string]interface{}{"Filter": ".device == 'iPhone'"}
	toRule := &blocks.Msg{Msg: ruleMsg, Route: "rule"}
	ch.InChan <- toRule

	outChan := make(chan *blocks.Msg)
	ch.AddChan <- &blocks.AddChanMsg{Route: "1", Channel: outChan}

	queryOutChan := make(blocks.MsgChan)
	ch.QueryChan <- &blocks.QueryMsg{MsgChan: queryOutChan, Route: "rule"}

	time.AfterFunc(time.Duration(5)*time.Second, func() {
		ch.QuitChan <- true
	})

	for {
		select {
		case messageI := <-queryOutChan:
			if !reflect.DeepEqual(messageI, ruleMsg) {
				c.Fail()
			}

		case message := <-outChan:
			log.Println(message)

		case err := <-ch.ErrChan:
			if err != nil {
				c.Errorf(err.Error())
			} else {
				return
			}
		}
	}
}

func (s *StreamSuite) TestMask(c *C) {
	log.Println("testing Mask")
	b, ch := newBlock("testingMask", "mask")
	go blocks.BlockRoutine(b)

	ruleMsg := map[string]interface{}{
		"Mask": map[string]interface{}{
			".foo": "{}",
		},
	}
	toRule := &blocks.Msg{Msg: ruleMsg, Route: "rule"}
	ch.InChan <- toRule

	outChan := make(chan *blocks.Msg)
	ch.AddChan <- &blocks.AddChanMsg{Route: "1", Channel: outChan}

	queryOutChan := make(blocks.MsgChan)
	ch.QueryChan <- &blocks.QueryMsg{MsgChan: queryOutChan, Route: "rule"}

	time.AfterFunc(time.Duration(5)*time.Second, func() {
		ch.QuitChan <- true
	})

	for {
		select {
		case messageI := <-queryOutChan:
			if !reflect.DeepEqual(messageI, ruleMsg) {
				log.Println("Rule mismatch:", messageI, ruleMsg)
				c.Fail()
			}

		case message := <-outChan:
			log.Println(message)

		case err := <-ch.ErrChan:
			if err != nil {
				c.Errorf(err.Error())
			} else {
				return
			}
		}
	}
}

func (s *StreamSuite) TestGetHTTPXML(c *C) {
	log.Println("testing GetHTTP with XML")
	b, ch := newBlock("testingGetHTTPXML", "gethttp")
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
func (s *StreamSuite) TestGetHTTP(c *C) {
	log.Println("testing GetHTTP")
	b, ch := newBlock("testingGetHTTP", "gethttp")
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

func (s *StreamSuite) TestFromHTTPStreamXML(c *C) {
	log.Println("testing FromHTTPStream with XML")
	b, ch := newBlock("testingFromHTTPStreamXML", "fromhttpstream")
	go blocks.BlockRoutine(b)
	outChan := make(chan *blocks.Msg)
	ch.AddChan <- &blocks.AddChanMsg{
		Route:   "out",
		Channel: outChan,
	}

	ruleMsg := map[string]interface{}{"Endpoint": "https://raw.github.com/nytlabs/streamtools/master/examples/odf.xml"}
	toRule := &blocks.Msg{Msg: ruleMsg, Route: "rule"}
	ch.InChan <- toRule

	queryOutChan := make(blocks.MsgChan)
	time.AfterFunc(time.Duration(1)*time.Second, func() {
		ch.QueryChan <- &blocks.QueryMsg{MsgChan: queryOutChan, Route: "rule"}
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
			message := messageI.(map[string]interface{})
			if !reflect.DeepEqual(message["Endpoint"], ruleMsg["Endpoint"]) {
				log.Println("Rule mismatch:", message["Endpoint"], ruleMsg["Endpoint"])
				c.Fail()
			}

		case messageI := <-outChan:
			message := messageI.Msg.(map[string]interface{})
			fmt.Printf("%s", message["data"])
		}
	}
}

func (s *StreamSuite) TestFromHTTPStream(c *C) {
	log.Println("testing FromHTTPStream")
	b, ch := newBlock("testingFromHTTPStream", "fromhttpstream")
	go blocks.BlockRoutine(b)
	outChan := make(chan *blocks.Msg)
	ch.AddChan <- &blocks.AddChanMsg{
		Route:   "out",
		Channel: outChan,
	}

	ruleMsg := map[string]interface{}{"Endpoint": "http://www.nytimes.com"}
	toRule := &blocks.Msg{Msg: ruleMsg, Route: "rule"}
	ch.InChan <- toRule

	queryOutChan := make(blocks.MsgChan)
	time.AfterFunc(time.Duration(1)*time.Second, func() {
		ch.QueryChan <- &blocks.QueryMsg{MsgChan: queryOutChan, Route: "rule"}
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
			message := messageI.(map[string]interface{})
			if !reflect.DeepEqual(message["Endpoint"], ruleMsg["Endpoint"]) {
				log.Println("Rule mismatch:", message["Endpoint"], ruleMsg["Endpoint"])
				c.Fail()
			}

		case <-outChan:
		}
	}
}

func (s *StreamSuite) TestFromFile(c *C) {
	log.Println("testing FromFile")
	b, ch := newBlock("testingFile", "fromfile")
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

func (s *StreamSuite) TestFromPost(c *C) {
	log.Println("testing FromPost")
	b, ch := newBlock("testingPost", "frompost")
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

func (s *StreamSuite) TestMap(c *C) {
	log.Println("testing Map")
	b, ch := newBlock("testingMap", "map")
	go blocks.BlockRoutine(b)

	outChan := make(chan *blocks.Msg)
	ch.AddChan <- &blocks.AddChanMsg{Route: "1", Channel: outChan}

	//outChan := make(chan *blocks.Msg)
	//ch.AddChan <- &blocks.AddChanMsg{
	//	Route:   "out",
	//	Channel: outChan,
	//}

	mapMsg := map[string]interface{}{"MegaBar": ".bar"}
	ruleMsg := map[string]interface{}{"Map": mapMsg, "Additive": false}

	// send rule
	toRule := &blocks.Msg{Msg: ruleMsg, Route: "rule"}
	ch.InChan <- toRule

	// send rule query in a sec
	queryOutChan := make(blocks.MsgChan)
	time.AfterFunc(time.Duration(1)*time.Second, func() {
		ch.QueryChan <- &blocks.QueryMsg{MsgChan: queryOutChan, Route: "rule"}
	})

	// quit after 5
	time.AfterFunc(time.Duration(5)*time.Second, func() {
		ch.QuitChan <- true
	})

	// send test input
	inputMsg := map[string]interface{}{"bar": "something", "foo": "another thing"}
	inputBlock := &blocks.Msg{Msg: inputMsg, Route: "in"}
	time.AfterFunc(time.Duration(2)*time.Second, func() {
		ch.InChan <- inputBlock
	})

	for {
		select {
		case messageI := <-queryOutChan:
			message := messageI.(map[string]interface{})
			if !reflect.DeepEqual(message["Map"], ruleMsg["Map"]) {
				log.Println(messageI)
				log.Println("was expecting", ruleMsg["Map"], "but got", message["Map"])
				c.Fail()
			}
		case err := <-ch.ErrChan:
			if err != nil {
				c.Errorf(err.Error())
			} else {
				return
			}
		case messageI := <-outChan:
			message := messageI.Msg.(map[string]interface{})
			log.Println(message)
			c.Assert(message["MegaBar"], Equals, "something")
			c.Assert(message["foo"], IsNil)
		}
	}
}

func (s *StreamSuite) TestHistogram(c *C) {
	log.Println("testing Histogram")
	b, ch := newBlock("testingHistogram", "histogram")
	go blocks.BlockRoutine(b)
	outChan := make(chan *blocks.Msg)
	ch.AddChan <- &blocks.AddChanMsg{
		Route:   "out",
		Channel: outChan,
	}

	ruleMsg := map[string]interface{}{"Window": "10s", "Path": ".data"}
	toRule := &blocks.Msg{Msg: ruleMsg, Route: "rule"}
	ch.InChan <- toRule

	queryOutChan := make(blocks.MsgChan)
	time.AfterFunc(time.Duration(1)*time.Second, func() {
		ch.QueryChan <- &blocks.QueryMsg{MsgChan: queryOutChan, Route: "rule"}
	})

	time.AfterFunc(time.Duration(5)*time.Second, func() {
		ch.QuitChan <- true
	})
	for {
		select {
		case messageI := <-queryOutChan:
			if !reflect.DeepEqual(messageI, ruleMsg) {
				c.Fail()
			}

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

func (s *StreamSuite) TestTimeseries(c *C) {
	log.Println("testing Timeseries")
	b, ch := newBlock("testingTimeseries", "timeseries")
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

func (s *StreamSuite) TestGaussian(c *C) {
	log.Println("testing Gaussian")
	b, ch := newBlock("testingGaussian", "gaussian")
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
		case <-outChan:
		}
	}
}

func (s *StreamSuite) TestZipf(c *C) {
	log.Println("testing Zipf")
	b, ch := newBlock("testingZipf", "zipf")
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
		case <-outChan:
		}
	}
}

func (s *StreamSuite) TestPoisson(c *C) {
	loghub.Start()
	log.Println("testing Poisson")
	b, ch := newBlock("testingPoisson", "poisson")
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
		case <-outChan:
		}
	}
}

func (s *StreamSuite) TestToElasticsearch(c *C) {
	log.Println("testing ToElasticsearch")
	b, ch := newBlock("testingToElasticsearch", "toelasticsearch")
	go blocks.BlockRoutine(b)
	outChan := make(chan *blocks.Msg)
	ch.AddChan <- &blocks.AddChanMsg{
		Route:   "out",
		Channel: outChan,
	}

	ruleMsg := map[string]interface{}{"Host": "localhost", "Port": "9200", "Index": "librarytest", "IndexType": "foobar"}
	rule := &blocks.Msg{Msg: ruleMsg, Route: "rule"}
	ch.InChan <- rule

	queryOutChan := make(blocks.MsgChan)
	time.AfterFunc(time.Duration(1)*time.Second, func() {
		ch.QueryChan <- &blocks.QueryMsg{MsgChan: queryOutChan, Route: "rule"}
	})

	time.AfterFunc(time.Duration(5)*time.Second, func() {
		ch.QuitChan <- true
	})
	for {
		select {
		case messageI := <-queryOutChan:
			if !reflect.DeepEqual(messageI, ruleMsg) {
				log.Println("rule mismatch:", messageI, ruleMsg)
				c.Fail()
			}

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

func (s *StreamSuite) TestUnpack(c *C) {
	loghub.Start()
	log.Println("testing unpack")
	b, ch := newBlock("testingunpack", "unpack")
	go blocks.BlockRoutine(b)
	outChan := make(chan *blocks.Msg)
	ch.AddChan <- &blocks.AddChanMsg{
		Route:   "out",
		Channel: outChan,
	}
	time.AfterFunc(time.Duration(5)*time.Second, func() {
		ch.QuitChan <- true
	})
	ruleMsg := map[string]interface{}{"Path": ".a"}
	rule := &blocks.Msg{Msg: ruleMsg, Route: "rule"}
	ch.InChan <- rule
	m := map[string]string{"b": "test"}
	arr := []interface{}{m}
	inMsg := map[string]interface{}{"a": arr}
	ch.InChan <- &blocks.Msg{Msg: inMsg, Route: "in"}
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

func (s *StreamSuite) TestMovingAverage(c *C) {
	loghub.Start()
	log.Println("testing moving average")
	b, ch := newBlock("testing movingaverave", "movingaverage")
	go blocks.BlockRoutine(b)
	outChan := make(chan *blocks.Msg)
	ch.AddChan <- &blocks.AddChanMsg{
		Route:   "out",
		Channel: outChan,
	}
	time.AfterFunc(time.Duration(5)*time.Second, func() {
		ch.QuitChan <- true
	})
	ruleMsg := map[string]interface{}{"Path": ".a", "Window": "4s"}
	rule := &blocks.Msg{Msg: ruleMsg, Route: "rule"}
	ch.InChan <- rule
	/*
		m := map[string]string{"b": "test"}
		arr := []interface{}{m}
		inMsg := map[string]interface{}{"a": arr}
		ch.InChan <- &blocks.Msg{Msg: inMsg, Route: "in"}
	*/
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

func (s *StreamSuite) TestPackByValue(c *C) {
	loghub.Start()
	log.Println("testing packbyvalue")
	b, ch := newBlock("testing packbyvalue", "packbyvalue")
	go blocks.BlockRoutine(b)
	outChan := make(chan *blocks.Msg)
	ch.AddChan <- &blocks.AddChanMsg{
		Route:   "out",
		Channel: outChan,
	}
	time.AfterFunc(time.Duration(5)*time.Second, func() {
		ch.QuitChan <- true
	})
	ruleMsg := map[string]interface{}{"Path": ".a", "EmitAfter": "4s"}
	rule := &blocks.Msg{Msg: ruleMsg, Route: "rule"}
	ch.InChan <- rule
	/*
		m := map[string]string{"b": "test"}
		arr := []interface{}{m}
		inMsg := map[string]interface{}{"a": arr}
		ch.InChan <- &blocks.Msg{Msg: inMsg, Route: "in"}
	*/
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

func (s *StreamSuite) TestSet(c *C) {
	loghub.Start()
	log.Println("testing set")
	b, ch := newBlock("testing set", "set")
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

func (s *StreamSuite) TestJoin(c *C) {
	loghub.Start()
	log.Println("testing join")
	b, ch := newBlock("testing join", "join")
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
		case <-outChan:
		}
	}
}

func (s *StreamSuite) TestParseXML(c *C) {
	log.Println("testing ParseXML")
	b, ch := newBlock("testingParseXML", "parsexml")
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

func (s *StreamSuite) TestFromPostXML(c *C) {
	log.Println("testing fromPost with XML")
	b, ch := newBlock("testingFromPostXML", "frompost")
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

func (s *StreamSuite) TestDeDupe(c *C) {
	loghub.Start()
	log.Println("testing dedupe")
	b, ch := newBlock("testing dedupe", "dedupe")

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

func (s *StreamSuite) TestCache(c *C) {
	loghub.Start()
	log.Println("testing cache")
	b, ch := newBlock("testing cache", "cache")

	go blocks.BlockRoutine(b)
	outChan := make(chan *blocks.Msg)
	ch.AddChan <- &blocks.AddChanMsg{
		Route:   "out",
		Channel: outChan,
	}
	ruleMsg := map[string]interface{}{"KeyPath": ".name", "ValuePath": ".count", "TimeToLive": "1m"}
	rule := &blocks.Msg{Msg: ruleMsg, Route: "rule"}
	ch.InChan <- rule

	// Add some data to the cache
	time.AfterFunc(time.Duration(1)*time.Second, func() {
		ch.InChan <- &blocks.Msg{Msg: map[string]interface{}{"count": "100", "name": "The New York Times"}, Route: "in"}
		ch.InChan <- &blocks.Msg{Msg: map[string]interface{}{"count": "4", "name": "The New York Times"}, Route: "in"}
		ch.InChan <- &blocks.Msg{Msg: map[string]interface{}{"count": "50", "name": "Hacks/Hackers"}, Route: "in"}
	})

	// Query for keys
	keysChan := make(blocks.MsgChan)
	time.AfterFunc(time.Duration(2)*time.Second, func() {
		ch.QueryChan <- &blocks.QueryMsg{MsgChan: keysChan, Route: "keys"}
	})

	// And values
	valuesChan := make(blocks.MsgChan)
	time.AfterFunc(time.Duration(2)*time.Second, func() {
		ch.QueryChan <- &blocks.QueryMsg{MsgChan: valuesChan, Route: "values"}
	})

	// And the entire cache contents
	dumpChan := make(blocks.MsgChan)
	time.AfterFunc(time.Duration(2)*time.Second, func() {
		ch.QueryChan <- &blocks.QueryMsg{MsgChan: dumpChan, Route: "dump"}
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

		case messageI := <-keysChan:
			message := messageI.(map[string]interface{})
			keys := message["keys"].([]string)
			c.Assert(keys, DeepEquals, []string{"The New York Times", "Hacks/Hackers"})

		case messageI := <-valuesChan:
			message := messageI.(map[string]interface{})
			values := message["values"].([]interface{})
			c.Assert(values, HasLen, 2)

		case messageI := <-dumpChan:
			message := messageI.(map[string]interface{})
			c.Assert(message["dump"], HasLen, 2)

		case messageI := <-outChan:
			message := messageI.Msg.(map[string]interface{})
			log.Println(message)
		}
	}
}
