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
}

func (b *inBlock) run(topic string) {
	// set block function going
	go b.f(b.inChan, b.ruleChan)
	// connect to NSQ
	go nsqReader(topic, b.inChan)
	// set the rule server going
	go b.listenForRules()
	// block until an os.signal
	<-b.sigChan
}

func NewInBlock(f inBlockRoutine) *inBlock {
	b := newBlock()
	inChan := make(chan simplejson.Json)
	return &inBlock{b, inChan, f}
}

type outBlockRoutine func(outChan chan simplejson.Json, ruleChan chan *http.Request)

type stateBlockRoutine func(inChan chan simplejson.Json, ruleChan chan *http.Request, readChan chan *http.Request)

type inOutBlockRoutine func(inChan chan simplejson.Json, outChan chan simplejson.Json, ruleChan chan *http.Request)
