package library

import (
	"github.com/nytlabs/streamtools/st/blocks" // blocks
	"log"
	"testing"
	"time"
)

func newBlock(id, kind string) (blocks.BlockInterface, blocks.BlockChans) {

	library := map[string]func() blocks.BlockInterface{
		"count":   NewCount,
		"toFile":  NewToFile,
		"fromNSQ": NewFromNSQ,
		"fromSQS": NewFromSQS,
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

func TestFromNSQ(t *testing.T) {
	log.Println("testing fromNSQ")

	b, c := newBlock("testingfromNSQ", "fromNSQ")
	go blocks.BlockRoutine(b)

	outChan := make(chan *blocks.Msg)
	c.AddChan <- &blocks.AddChanMsg{Route: "1", Channel: outChan}

	fooMsg := map[string]interface{}{"ReadTopic": "test", "LookupdAddr": "127.0.0.1:4161", "ReadChannel": "nsq_to_file", "MaxInFlight": 100}
	rule := &blocks.Msg{Msg: fooMsg, Route: "rule"}
	c.InChan <- rule

	time.AfterFunc(time.Duration(5)*time.Second, func() {
		c.QuitChan <- true
	})

	for {
		select {
		case message := <-outChan:
			log.Println("caught message on outChan")
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

func TestCount(t *testing.T) {
	log.Println("testing Count")
	b, c := newBlock("testingCount", "count")
	go blocks.BlockRoutine(b)
	time.AfterFunc(time.Duration(5)*time.Second, func() {
		c.QuitChan <- true
	})
	err := <-c.ErrChan
	if err != nil {
		t.Errorf(err.Error())
	}
}

func TestToFile(t *testing.T) {
	log.Println("testing toFile")
	b, c := newBlock("testingToFile", "toFile")
	go blocks.BlockRoutine(b)
	time.AfterFunc(time.Duration(5)*time.Second, func() {
		c.QuitChan <- true
	})
	err := <-c.ErrChan
	if err != nil {
		t.Errorf(err.Error())
	}
}

func TestFromSQS(t *testing.T) {
	log.Println("testing FromSQS")
	b, c := newBlock("testingFromSQS", "fromSQS")
	go blocks.BlockRoutine(b)
	time.AfterFunc(time.Duration(5)*time.Second, func() {
		c.QuitChan <- true
	})
	err := <-c.ErrChan
	if err != nil {
		t.Errorf(err.Error())
	}
}
