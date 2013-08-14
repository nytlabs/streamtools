package streamtools

import (
	"fmt"
	"github.com/bitly/go-simplejson"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

// generic block

type block struct {
	ruleChan chan *simplejson.Json
	sigChan  chan os.Signal
}

func (b *block) updateRule(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		log.Fatalf(err.Error())
	}
	rule, err := simplejson.NewJson(body)
	if err != nil {
		log.Fatalf(err.Error())
	}
	b.ruleChan <- rule
	fmt.Fprintf(w, "thanks buddy")
}

func (b *block) listenForRules() {
	// listen for rule changes
	http.HandleFunc("/", b.updateRule)
	http.ListenAndServe(":8080", nil)
}

func newBlock() *block {
	return &block{
		ruleChan: make(chan *simplejson.Json),
		sigChan:  make(chan os.Signal),
	}
}

// inBlocks only have an input from stream tools

type inBlockRoutine func(inChan chan simplejson.Json, ruleChan chan *simplejson.Json)

type inBlock struct {
	*block // embeds the block type, giving us updateRule and the sigChan for free
	inChan chan simplejson.Json
	f      inBlockRoutine
	name   string
}

func (b *inBlock) Run(topic string) {
	// set block function going
	go b.f(b.inChan, b.ruleChan)
	// connect to NSQ
	go nsqReader(topic, b.name, b.inChan)
	// set the rule server going
	go b.listenForRules()
	// block until an os.signal
	<-b.sigChan
}

func NewInBlock(f inBlockRoutine, name string) *inBlock {
	b := newBlock()
	inChan := make(chan simplejson.Json)
	return &inBlock{b, inChan, f, name}
}

// outBlocks only have an output to streamtools

type outBlockRoutine func(outChan chan simplejson.Json, ruleChan chan *simplejson.Json)

type outBlock struct {
	*block  // embeds the block type, giving us updateRule and the sigChan for free
	outChan chan simplejson.Json
	f       outBlockRoutine
	name    string
}

func (b *outBlock) Run(topic string) {
	// set block function going
	go b.f(b.outChan, b.ruleChan)
	// connect to NSQ
	log.Println("starting topic:", topic, "with channel:", b.name)
	go nsqWriter(topic, b.name, b.outChan)
	// set the rule server going
	go b.listenForRules()
	// block until an os.signal
	<-b.sigChan
}

func NewOutBlock(f outBlockRoutine, name string) *outBlock {
	b := newBlock()
	outChan := make(chan simplejson.Json)
	return &outBlock{b, outChan, f, name}
}

//

type stateBlockRoutine func(inChan chan simplejson.Json, ruleChan chan *http.Request, readChan chan *http.Request)

type inOutBlockRoutine func(inChan chan simplejson.Json, outChan chan simplejson.Json, ruleChan chan *http.Request)
