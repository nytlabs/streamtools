package library

import (
	"github.com/nytlabs/streamtools/st/blocks" // blocks
	"github.com/nytlabs/streamtools/st/loghub"
	"log"
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

func TestToFromNSQ(t *testing.T) {
	loghub.Start()
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
				t.Errorf(err.Error())
			} else {
				return
			}
		case err := <-fromC.ErrChan:
			if err != nil {
				t.Errorf(err.Error())
			} else {
				return
			}
		}
	}
}

func TestCount(t *testing.T) {
	loghub.Start()

	log.Println("testing Count")
	b, c := newBlock("testingCount", "count")
	go blocks.BlockRoutine(b)

	ruleMsg := map[string]interface{}{"Window": "1s"}
	toRule := &blocks.Msg{Msg: ruleMsg, Route: "rule"}
	c.InChan <- toRule

	outChan := make(chan *blocks.Msg)
	c.AddChan <- &blocks.AddChanMsg{Route: "1", Channel: outChan}

	queryOutChan := make(chan interface{})
	c.QueryChan <- &blocks.QueryMsg{RespChan: queryOutChan, Route: "rule"}

	time.AfterFunc(time.Duration(5)*time.Second, func() {
		c.QuitChan <- true
	})

	for {
		select {
		case messageI := <-queryOutChan:
			if !reflect.DeepEqual(messageI, ruleMsg) {
				t.Fail()
			}

		case message := <-outChan:
			log.Println(message)

		case err := <-c.ErrChan:
			if err != nil {
				t.Errorf(err.Error())
			} else {
				return
			}
		}
	}
}

func TestToFile(t *testing.T) {
	loghub.Start()
	log.Println("testing toFile")
	b, c := newBlock("testingToFile", "toFile")
	go blocks.BlockRoutine(b)

	ruleMsg := map[string]interface{}{"Filename": "foobar.log"}
	toRule := &blocks.Msg{Msg: ruleMsg, Route: "rule"}
	c.InChan <- toRule

	outChan := make(chan *blocks.Msg)
	c.AddChan <- &blocks.AddChanMsg{Route: "1", Channel: outChan}

	queryOutChan := make(chan interface{})
	c.QueryChan <- &blocks.QueryMsg{RespChan: queryOutChan, Route: "rule"}

	time.AfterFunc(time.Duration(5)*time.Second, func() {
		c.QuitChan <- true
	})

	for {
		select {
		case messageI := <-queryOutChan:
			if !reflect.DeepEqual(messageI, ruleMsg) {
				t.Fail()
			}

		case message := <-outChan:
			log.Println(message)

		case err := <-c.ErrChan:
			if err != nil {
				t.Errorf(err.Error())
			} else {
				return
			}
		}
	}
}

func TestFromSQS(t *testing.T) {
	loghub.Start()
	log.Println("testing FromSQS")
	b, c := newBlock("testingFromSQS", "fromSQS")
	go blocks.BlockRoutine(b)

	ruleMsg := map[string]interface{}{"SQSEndpoint": "foobarbaz", "AccessKey": "123access", "AccessSecret": "123secret"}
	toRule := &blocks.Msg{Msg: ruleMsg, Route: "rule"}
	c.InChan <- toRule

	outChan := make(chan *blocks.Msg)
	c.AddChan <- &blocks.AddChanMsg{Route: "1", Channel: outChan}

	queryOutChan := make(chan interface{})
	c.QueryChan <- &blocks.QueryMsg{RespChan: queryOutChan, Route: "rule"}

	time.AfterFunc(time.Duration(5)*time.Second, func() {
		c.QuitChan <- true
	})

	for {
		select {
		case messageI := <-queryOutChan:
			if !reflect.DeepEqual(messageI, ruleMsg) {
				t.Fail()
			}

		case message := <-outChan:
			log.Println(message)

		case err := <-c.ErrChan:
			if err != nil {
				t.Errorf(err.Error())
			} else {
				return
			}
		}
	}
}

func TestSync(t *testing.T) {
	loghub.Start()
	log.Println("testing Sync")
	b, c := newBlock("testingSync", "sync")
	go blocks.BlockRoutine(b)
	time.AfterFunc(time.Duration(5)*time.Second, func() {
		c.QuitChan <- true
	})
	err := <-c.ErrChan
	if err != nil {
		t.Errorf(err.Error())
	}
}

func TestTicker(t *testing.T) {
	loghub.Start()
	log.Println("testing Ticker")
	b, c := newBlock("testingTicker", "ticker")
	go blocks.BlockRoutine(b)
	outChan := make(chan *blocks.Msg)
	c.AddChan <- &blocks.AddChanMsg{
		Route:   "out",
		Channel: outChan,
	}
	time.AfterFunc(time.Duration(5)*time.Second, func() {
		c.QuitChan <- true
	})
	for {
		select {
		case err := <-c.ErrChan:
			if err != nil {
				t.Errorf(err.Error())
			} else {
				return
			}
		case <-outChan:
		}
	}
}

func TestFilter(t *testing.T) {
	loghub.Start()
	log.Println("testing Filter")
	b, c := newBlock("testingFilter", "filter")
	go blocks.BlockRoutine(b)

	ruleMsg := map[string]interface{}{"Filter": ".device == 'iPhone'"}
	toRule := &blocks.Msg{Msg: ruleMsg, Route: "rule"}
	c.InChan <- toRule

	outChan := make(chan *blocks.Msg)
	c.AddChan <- &blocks.AddChanMsg{Route: "1", Channel: outChan}

	queryOutChan := make(chan interface{})
	c.QueryChan <- &blocks.QueryMsg{RespChan: queryOutChan, Route: "rule"}

	time.AfterFunc(time.Duration(5)*time.Second, func() {
		c.QuitChan <- true
	})

	for {
		select {
		case messageI := <-queryOutChan:
			if !reflect.DeepEqual(messageI, ruleMsg) {
				t.Fail()
			}

		case message := <-outChan:
			log.Println(message)

		case err := <-c.ErrChan:
			if err != nil {
				t.Errorf(err.Error())
			} else {
				return
			}
		}
	}
}

func TestMask(t *testing.T) {
	loghub.Start()
	log.Println("testing Mask")
	b, c := newBlock("testingMask", "mask")
	go blocks.BlockRoutine(b)

	ruleMsg := map[string]interface{}{"Mask": "{}"}
	toRule := &blocks.Msg{Msg: ruleMsg, Route: "rule"}
	c.InChan <- toRule

	outChan := make(chan *blocks.Msg)
	c.AddChan <- &blocks.AddChanMsg{Route: "1", Channel: outChan}

	queryOutChan := make(chan interface{})
	c.QueryChan <- &blocks.QueryMsg{RespChan: queryOutChan, Route: "rule"}

	time.AfterFunc(time.Duration(5)*time.Second, func() {
		c.QuitChan <- true
	})

	for {
		select {
		case messageI := <-queryOutChan:
			if !reflect.DeepEqual(messageI, ruleMsg) {
				t.Fail()
			}

		case message := <-outChan:
			log.Println(message)

		case err := <-c.ErrChan:
			if err != nil {
				t.Errorf(err.Error())
			} else {
				return
			}
		}
	}
}

func TestGetHTTP(t *testing.T) {
	loghub.Start()
	log.Println("testing GetHTTP")
	b, c := newBlock("testingGetHTTP", "getHTTP")
	go blocks.BlockRoutine(b)
	outChan := make(chan *blocks.Msg)
	c.AddChan <- &blocks.AddChanMsg{
		Route:   "out",
		Channel: outChan,
	}
	time.AfterFunc(time.Duration(5)*time.Second, func() {
		c.QuitChan <- true
	})
	for {
		select {
		case err := <-c.ErrChan:
			if err != nil {
				t.Errorf(err.Error())
			} else {
				return
			}
		case <-outChan:
		}
	}
}

func TestFromHTTPStream(t *testing.T) {
	loghub.Start()
	log.Println("testing FromHTTPStream")
	b, c := newBlock("testingFromHTTPStream", "fromHTTPStream")
	go blocks.BlockRoutine(b)
	outChan := make(chan *blocks.Msg)
	c.AddChan <- &blocks.AddChanMsg{
		Route:   "out",
		Channel: outChan,
	}

	ruleMsg := map[string]interface{}{"Endpoint": "http://www.nytimes.com"}
	toRule := &blocks.Msg{Msg: ruleMsg, Route: "rule"}
	c.InChan <- toRule

	time.AfterFunc(time.Duration(5)*time.Second, func() {
		c.QuitChan <- true
	})

	for {
		select {
		case err := <-c.ErrChan:
			if err != nil {
				t.Errorf(err.Error())
			} else {
				return
			}
		case <-outChan:
		}
	}
}

func TestFromPost(t *testing.T) {
	loghub.Start()
	log.Println("testing FromPost")
	b, c := newBlock("testingPst", "fromPost")
	go blocks.BlockRoutine(b)
	outChan := make(chan *blocks.Msg)
	c.AddChan <- &blocks.AddChanMsg{
		Route:   "out",
		Channel: outChan,
	}

	inputMsg := map[string]interface{}{"Foo": "BAR"}
	inputBlock := &blocks.Msg{Msg: inputMsg, Route: "in"}
	c.InChan <- inputBlock

	time.AfterFunc(time.Duration(5)*time.Second, func() {
		c.QuitChan <- true
	})
	for {
		select {
		case err := <-c.ErrChan:
			if err != nil {
				t.Errorf(err.Error())
			} else {
				return
			}
		case <-outChan:
		}
	}
}

func TestMap(t *testing.T) {
	loghub.Start()
	log.Println("testing Map")
	b, c := newBlock("testingMap", "map")
	go blocks.BlockRoutine(b)
	outChan := make(chan *blocks.Msg)
	c.AddChan <- &blocks.AddChanMsg{
		Route:   "out",
		Channel: outChan,
	}

	mapMsg := map[string]interface{}{"Foo": ".bar"}
	ruleMsg := map[string]interface{}{"Map": mapMsg}
	toRule := &blocks.Msg{Msg: ruleMsg, Route: "rule"}
	c.InChan <- toRule

	queryOutChan := make(chan interface{})
	c.QueryChan <- &blocks.QueryMsg{RespChan: queryOutChan, Route: "rule"}

	time.AfterFunc(time.Duration(5)*time.Second, func() {
		c.QuitChan <- true
	})
	for {
		select {
		case messageI := <-queryOutChan:
			message := messageI.(map[string]interface{})
			if !reflect.DeepEqual(message["Map"], ruleMsg["Map"]) {
				t.Fail()
			}
		case err := <-c.ErrChan:
			if err != nil {
				t.Errorf(err.Error())
			} else {
				return
			}
		case <-outChan:
		}
	}
}

func TestHistogram(t *testing.T) {
	loghub.Start()
	log.Println("testing Histogram")
	b, c := newBlock("testingHistogram", "histogram")
	go blocks.BlockRoutine(b)
	outChan := make(chan *blocks.Msg)
	c.AddChan <- &blocks.AddChanMsg{
		Route:   "out",
		Channel: outChan,
	}

	ruleMsg := map[string]interface{}{"Window": "10s", "Path": ".data"}
	toRule := &blocks.Msg{Msg: ruleMsg, Route: "rule"}
	c.InChan <- toRule

	queryOutChan := make(chan interface{})
	c.QueryChan <- &blocks.QueryMsg{RespChan: queryOutChan, Route: "rule"}

	time.AfterFunc(time.Duration(5)*time.Second, func() {
		c.QuitChan <- true
	})
	for {
		select {
		case messageI := <-queryOutChan:
			if !reflect.DeepEqual(messageI, ruleMsg) {
				t.Fail()
			}

		case err := <-c.ErrChan:
			if err != nil {
				t.Errorf(err.Error())
			} else {
				return
			}
		case <-outChan:
		}
	}
}

func TestTimeseries(t *testing.T) {
	log.Println("testing Timeseries")
	b, c := newBlock("testingTimeseries", "timeseries")
	go blocks.BlockRoutine(b)
	outChan := make(chan *blocks.Msg)
	c.AddChan <- &blocks.AddChanMsg{
		Route:   "out",
		Channel: outChan,
	}
	time.AfterFunc(time.Duration(5)*time.Second, func() {
		c.QuitChan <- true
	})
	for {
		select {
		case err := <-c.ErrChan:
			if err != nil {
				t.Errorf(err.Error())
			} else {
				return
			}
		case <-outChan:
		}
	}
}
