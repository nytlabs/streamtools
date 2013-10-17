package streamtools

import (
	"github.com/bitly/go-simplejson"
	"log"
)

// Block is the basic interface for processing units in streamtools
type Block interface {
	// blockRoutine is the central processing routine for a block. All the work gets done in here
	blockRoutine()
	// init routines TODO should just be init
	initOutChans()
	// a set of accessors are provided so that a block creator can access certain aspects of a block
	getID() string
	getBlockType() string
	getInChan() chan *simplejson.Json
	getOutChans() map[string]chan *simplejson.Json
	getRouteChan(string) chan routeResponse
	getRoutes() []string
	// some aspects of a block can also be set by the block creator
	setInChan(chan *simplejson.Json)
	setID(string)
	setOutChan(string, chan *simplejson.Json)
	createOutChan(string) chan *simplejson.Json
}

// The AbstractBlock struct defines the attributes a block must have
type AbstractBlock struct {
	// the ID is the unique key by which streamtools refers to the block
	ID string
	// blockType defines what kind of block this
	blockType string
	// the inChan passes messages from elsewhere into this block
	inChan chan *simplejson.Json
	// the outChan sends messages from this block elsewhere
	outChans map[string]chan *simplejson.Json
	// the routes map is used to define arbitrary streamtools endpoints for this block
	routes map[string]chan routeResponse
}

// SIMPLE GETTERS AND SETTERS

func (self *AbstractBlock) getID() string {
	return self.ID
}

func (self *AbstractBlock) getBlockType() string {
	return self.blockType
}

func (self *AbstractBlock) getInChan() chan *simplejson.Json {
	return self.inChan
}

func (self *AbstractBlock) getOutChans() map[string]chan *simplejson.Json {
	return self.outChans
}

// ROUTES

// returns a channel specified by an endpoint name
func (self *AbstractBlock) getRouteChan(name string) chan routeResponse {
	if val, ok := self.routes[name]; ok {
		return val
	}
	// TODO return a proper error on this if key is not found.
	return nil
}

// getRoutes returns all of the route names specified by the block
func (self *AbstractBlock) getRoutes() []string {
	routeNames := make([]string, len(self.routes))
	i := 0
	for name, _ := range self.routes {
		routeNames[i] = name
		i += 1
	}
	return routeNames
}

func (self *AbstractBlock) setInChan(inChan chan *simplejson.Json) {
	log.Println("setting in block of", self.ID, "to", inChan)
	self.inChan = inChan
}

func (self *AbstractBlock) setID(id string) {
	self.ID = id
}

func (self *AbstractBlock) setOutChan(toBlockID string, outChan chan *simplejson.Json) {
	self.outChans[toBlockID] = outChan
	log.Println(self.ID, "'s out channels are now:", self.outChans)
}

func (self *AbstractBlock) createOutChan(toBlockID string) chan *simplejson.Json {
	log.Println("setting out channel of", self.ID)
	outChan := make(chan *simplejson.Json)
	self.outChans[toBlockID] = outChan
	log.Println(self.ID, "'s out channels are now:", self.outChans)
	return outChan
}

func (self *AbstractBlock) initOutChans() {
	self.outChans = make(map[string]chan *simplejson.Json)
}

// NewBlock returns a block from the block factory specified in the library
func NewBlock(blockType string) Block {
	blockTemplate, ok := library[blockType]
	if !ok {
		log.Fatal("couldn't find block", blockType)
	}
	block := blockTemplate.blockFactory()
	block.initOutChans()
	return block
}
