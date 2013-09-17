/*
Package streamtools provides a set of tools for working with streams of JSON
*/
package streamtools

import (
	"fmt"
	"github.com/bitly/go-simplejson"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

// the block type defines the basic structure of a block
type block struct {
	RuleChan chan *simplejson.Json
	sigChan  chan os.Signal
	name     string
}

// updateRule passes the new rule into the block
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
	b.RuleChan <- rule
}

// listen for rule changes
func (b *block) listenForRules() {
	http.HandleFunc("/rules", b.updateRule)
}

// StartServer simply sets the block's http server listening on the specified port
func (b *block) StartServer(port string) {
	http.ListenAndServe(":"+port, nil)
}

// newBlock returns a generic streamtools block
func newBlock(name string) *block {
	return &block{
		RuleChan: make(chan *simplejson.Json, 1),
		sigChan:  make(chan os.Signal),
		name:     name,
	}
}





// inBlocks only have an input from stream tools
type inBlock struct {
	*block // embeds the block type, giving us updateRule and the sigChan for free
	inChan chan *simplejson.Json
	f      inBlockRoutine
}

// implement an inBlockRoutine to define a new inBlock
type inBlockRoutine func(inChan chan *simplejson.Json, RuleChan chan *simplejson.Json)

/*
Run does everything necessary to set the inBlock running.

It starts the inBlockRoutine, which recives messages on the block's inChan channel. It also
connects to the NSQ service, and starts listening on the block's port for new rules. Send this
block an os.Signal to stop it. 
*/
func (b *inBlock) Run(topic string, port string) {
	// set block function going
	go b.f(b.inChan, b.RuleChan)
	// connect to NSQ
	go nsqReader(topic, b.name, b.inChan)
	// set the rule server going
	b.listenForRules()
	go b.StartServer(port)
	// block until an os.signal
	<-b.sigChan
}

/*
NewInBlock creates and initialises an InBlock. It contains, in addition to the generic block, 
an inChan channel which is shared with the supplied inBlockRoutine. 
*/
func NewInBlock(f inBlockRoutine, name string) *inBlock {
	b := newBlock(name)
	inChan := make(chan *simplejson.Json)
	return &inBlock{b, inChan, f}
}






// outBlocks only have an output to streamtools
type outBlock struct {
	*block  // embeds the block type, giving us updateRule and the sigChan for free
	outChan chan *simplejson.Json
	f       outBlockRoutine
}

// implement an outBlockRoutine to define a new outBlock
type outBlockRoutine func(outChan chan *simplejson.Json, RuleChan chan *simplejson.Json)

/*
Run does everything necessary to set the outblock running.

It starts the outBlockRoutine, which sends messages to the block's outChan channel. 

It also connects to the NSQ service, and starts listening on the block's port for new rules. Send this
block an os.Signal to stop it. 
*/
func (b *outBlock) Run(topic string, port string) {
	// set block function going
	go b.f(b.outChan, b.RuleChan)
	// connect to NSQ
	go nsqWriter(topic, b.outChan)
	// set the rule server going
	b.listenForRules()
	go b.StartServer(port)
	// block until an os.signal
	<-b.sigChan
}

/*
NewOutBlock creates and initialises an outBlock. It contains, in addition to the generic block, 
an outChan channel which is shared with the supplied outBlockRoutine. 
*/
func NewOutBlock(f outBlockRoutine, name string) *outBlock {
	b := newBlock(name)
	outChan := make(chan *simplejson.Json)
	return &outBlock{b, outChan, f}
}





// state blocks only have inbound data, but maintain a state which you can query via HTTP
type stateBlock struct {
	*block
	inChan    chan *simplejson.Json
	queryChan chan stateQuery // for requests to query the state
	f         stateBlockRoutine
}

// implement an stateBlockRoutine to define a new stateBlock
type stateBlockRoutine func(inChan chan *simplejson.Json, RuleChan chan *simplejson.Json, queryChan chan stateQuery)

/*
Run does everything necessary to set the stateBlock running.

It starts the supplied stateBlockRoutine, which recieves messages on the block's inChan channel, and creates a 
handler on the block's server, accessed via /state, that returns JSON describing the block's current state. 

It also connects to the NSQ service, and starts listening on the block's port for new rules. Send this
block an os.Signal to stop it. 
*/
func (b *stateBlock) Run(topic string, port string) {
	go b.f(b.inChan, b.RuleChan, b.queryChan)
	go nsqReader(topic, b.name, b.inChan)
	b.listenForStateQuery()
	b.listenForRules()
	go b.StartServer(port)
	<-b.sigChan
}

// defines the form of a query for the state blokck
type stateQuery struct {
	responseChan chan *simplejson.Json
}

// query handler, called when an HTTP request for this block's state is made
func (b *stateBlock) queryState(w http.ResponseWriter, r *http.Request) {
	q := stateQuery{
		responseChan: make(chan *simplejson.Json),
	}
	b.queryChan <- q
	// block until the response
	response := <-q.responseChan
	msg, err := response.Encode()
	if err != nil {
		log.Fatalf(err.Error())
	}
	fmt.Fprintf(w, string(msg))
}

// listenForStateQuery adds the /state handler to this block's server
func (b *stateBlock) listenForStateQuery() {
	http.HandleFunc("/state", b.queryState)
}

/*
NewStateBlock creates and initialises a stateBlock. It contains, in addition to the generic block, 
the inChan channel and the queryChan channel, both of which are shared with the supplied stateBlockRoutine. 
*/
func NewStateBlock(f stateBlockRoutine, name string) *stateBlock {
	b := newBlock(name)
	inChan := make(chan *simplejson.Json)
	queryChan := make(chan stateQuery)
	return &stateBlock{b, inChan, queryChan, f}
}



// transfer blocks have both inbound and outbound data
type transferBlock struct {
	*block
	inChan  chan *simplejson.Json
	outChan chan *simplejson.Json
	f       transferBlockRoutine
}

// implement an transferBlockRoutine to define a new transferBlock
type transferBlockRoutine func(inChan chan *simplejson.Json, outChan chan *simplejson.Json, RuleChan chan *simplejson.Json)

/*
Run does everything necessary to set the transferBlock running.

It starts the supplied transferBlockRoutine, which recieves messages on the block's inChan channel and writes
messages on the block's outChan channel. 

It also connects to the NSQ service, and starts listening on the block's port for new rules. Send this
block an os.Signal to stop it. 
*/
func (b *transferBlock) Run(inTopic string, outTopic string, port string) {
	go b.f(b.inChan, b.outChan, b.RuleChan)
	go nsqReader(inTopic, b.name, b.inChan)
	go nsqWriter(outTopic, b.outChan)
	b.listenForRules()
	go b.StartServer(port)
	<-b.sigChan
}

/*
NewTransferBlock creates and initialises a transfeBlock. It contains, in addition to the generic block, 
the inChan channel and the queryChan channel, both of which are shared with the supplied stateBlockRoutine. 
*/
func NewTransferBlock(f transferBlockRoutine, name string) *transferBlock {
	b := newBlock(name)
	inChan := make(chan *simplejson.Json)
	outChan := make(chan *simplejson.Json)
	return &transferBlock{b, inChan, outChan, f}
}

// a seperate type of run for the multiple outputs
func (b *transferBlock) DeMuxRun(inTopic string, port string) {
	go b.f(b.inChan, b.outChan, b.RuleChan)
	go nsqReader(inTopic, b.name, b.inChan)
	go deMuxWriter(b.outChan)
	b.listenForRules()
	go b.StartServer(port)
	<-b.sigChan

}
