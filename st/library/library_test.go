package library

import (
	"github.com/nytlabs/streamtools/st/blocks" // blocks
	"github.com/nytlabs/streamtools/st/loghub"
	. "launchpad.net/gocheck"
	"log"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"
)

func newBlock(id, kind string) (blocks.BlockInterface, blocks.BlockChans) {

	library := map[string]func() blocks.BlockInterface{
		"count":          NewCount,
		"toFile":         NewToFile,
		"fromNSQ":        NewFromNSQ,
		"toNSQ":          NewToNSQ,
		"fromSQS":        NewFromSQS,
		"ticker":         NewTicker,
		"filter":         NewFilter,
		"mask":           NewMask,
		"fromHTTPStream": NewFromHTTPStream,
		"getHTTP":        NewGetHTTP,
		"sync":           NewSync,
		"fromPost":       NewFromPost,
		"map":            NewMap,
		"histogram":      NewHistogram,
		"timeseries":     NewTimeseries,
	}

	chans := blocks.BlockChans{
		InChan:    make(chan *blocks.Msg),
		QueryChan: make(chan *blocks.QueryMsg),
		AddChan:   make(chan *blocks.AddChanMsg),
		DelChan:   make(chan *blocks.Msg),
		ErrChan:   make(chan error),
		QuitChan:  make(chan bool),
	}

	// actual block
	b := library[kind]()
	b.Build(chans)

	return b, chans

}

// Hook up gocheck into the "go test" runner.
//func (s *StreamSuite) Test(c *C) { TestingT(t) }
func Test(t *testing.T) { TestingT(t) }

type StreamSuite struct{}

func (s *StreamSuite) SetUpSuite(c *C) {
	loghub.Start()
}

var _ = Suite(&StreamSuite{})

func (s *StreamSuite) TestHelloWorld(c *C) {
	log.Println("testing toNSQ")

	toB, toC := newBlock("testingToNSQ", "toNSQ")
	go blocks.BlockRoutine(toB)

	ruleMsg := map[string]interface{}{"Topic": "librarytest", "NsqdTCPAddrs": "127.0.0.1:4150"}
	toRule := &blocks.Msg{Msg: ruleMsg, Route: "rule"}
	toC.InChan <- toRule

	nsqMsg := map[string]interface{}{"Foo": "Bar"}
	postData := &blocks.Msg{Msg: nsqMsg, Route: "in"}
	toC.InChan <- postData

	time.AfterFunc(time.Duration(5)*time.Second, func() {
		toC.QuitChan <- true
	})

	log.Println("testing fromNSQ")

	fromB, fromC := newBlock("testingfromNSQ", "fromNSQ")
	go blocks.BlockRoutine(fromB)

	outChan := make(chan *blocks.Msg)
	fromC.AddChan <- &blocks.AddChanMsg{Route: "1", Channel: outChan}

	nsqSetup := map[string]interface{}{"ReadTopic": "librarytest", "LookupdAddr": "127.0.0.1:4161", "ReadChannel": "libtestchannel", "MaxInFlight": 100}
	fromRule := &blocks.Msg{Msg: nsqSetup, Route: "rule"}
	fromC.InChan <- fromRule

	time.AfterFunc(time.Duration(5)*time.Second, func() {
		fromC.QuitChan <- true
	})

	for {
		select {
		case message := <-outChan:
			log.Println(message)

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

	outChan := make(chan *blocks.Msg)
	ch.AddChan <- &blocks.AddChanMsg{Route: "1", Channel: outChan}

	queryOutChan := make(chan interface{})
	ch.QueryChan <- &blocks.QueryMsg{RespChan: queryOutChan, Route: "rule"}

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

func (s *StreamSuite) TestToFile(c *C) {
	log.Println("testing toFile")
	b, ch := newBlock("testingToFile", "toFile")
	go blocks.BlockRoutine(b)

	ruleMsg := map[string]interface{}{"Filename": "foobar.log"}
	toRule := &blocks.Msg{Msg: ruleMsg, Route: "rule"}
	ch.InChan <- toRule

	outChan := make(chan *blocks.Msg)
	ch.AddChan <- &blocks.AddChanMsg{Route: "1", Channel: outChan}

	queryOutChan := make(chan interface{})
	ch.QueryChan <- &blocks.QueryMsg{RespChan: queryOutChan, Route: "rule"}

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

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	}))
	defer ts.Close()

	b, ch := newBlock("testingFromSQS", "fromSQS")
	go blocks.BlockRoutine(b)

	ruleMsg := map[string]interface{}{"SQSEndpoint": ts.URL, "AccessKey": "123access", "AccessSecret": "123secret"}
	toRule := &blocks.Msg{Msg: ruleMsg, Route: "rule"}
	ch.InChan <- toRule

	outChan := make(chan *blocks.Msg)
	ch.AddChan <- &blocks.AddChanMsg{Route: "1", Channel: outChan}

	queryOutChan := make(chan interface{})
	ch.QueryChan <- &blocks.QueryMsg{RespChan: queryOutChan, Route: "rule"}

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

func (s *StreamSuite) TestFilter(c *C) {
	log.Println("testing Filter")
	b, ch := newBlock("testingFilter", "filter")
	go blocks.BlockRoutine(b)

	ruleMsg := map[string]interface{}{"Filter": ".device == 'iPhone'"}
	toRule := &blocks.Msg{Msg: ruleMsg, Route: "rule"}
	ch.InChan <- toRule

	outChan := make(chan *blocks.Msg)
	ch.AddChan <- &blocks.AddChanMsg{Route: "1", Channel: outChan}

	queryOutChan := make(chan interface{})
	ch.QueryChan <- &blocks.QueryMsg{RespChan: queryOutChan, Route: "rule"}

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

	ruleMsg := map[string]interface{}{"Mask": "{}"}
	toRule := &blocks.Msg{Msg: ruleMsg, Route: "rule"}
	ch.InChan <- toRule

	outChan := make(chan *blocks.Msg)
	ch.AddChan <- &blocks.AddChanMsg{Route: "1", Channel: outChan}

	queryOutChan := make(chan interface{})
	ch.QueryChan <- &blocks.QueryMsg{RespChan: queryOutChan, Route: "rule"}

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

func (s *StreamSuite) TestGetHTTP(c *C) {
	log.Println("testing GetHTTP")
	b, ch := newBlock("testingGetHTTP", "getHTTP")
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

func (s *StreamSuite) TestFromHTTPStream(c *C) {
	log.Println("testing FromHTTPStream")
	b, ch := newBlock("testingFromHTTPStream", "fromHTTPStream")
	go blocks.BlockRoutine(b)
	outChan := make(chan *blocks.Msg)
	ch.AddChan <- &blocks.AddChanMsg{
		Route:   "out",
		Channel: outChan,
	}

	ruleMsg := map[string]interface{}{"Endpoint": "http://www.nytimes.com"}
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
	b, ch := newBlock("testingPst", "fromPost")
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
	ch.AddChan <- &blocks.AddChanMsg{
		Route:   "out",
		Channel: outChan,
	}

	mapMsg := map[string]interface{}{"Foo": ".bar"}
	ruleMsg := map[string]interface{}{"Map": mapMsg}
	toRule := &blocks.Msg{Msg: ruleMsg, Route: "rule"}
	ch.InChan <- toRule

	queryOutChan := make(chan interface{})
	ch.QueryChan <- &blocks.QueryMsg{RespChan: queryOutChan, Route: "rule"}

	time.AfterFunc(time.Duration(5)*time.Second, func() {
		ch.QuitChan <- true
	})
	for {
		select {
		case messageI := <-queryOutChan:
			message := messageI.(map[string]interface{})
			if !reflect.DeepEqual(message["Map"], ruleMsg["Map"]) {
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

	queryOutChan := make(chan interface{})
	ch.QueryChan <- &blocks.QueryMsg{RespChan: queryOutChan, Route: "rule"}

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

func TestTimeseries(t *testing.T) {
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
				t.Errorf(err.Error())
			} else {
				return
			}
		case <-outChan:
		}
	}
}
