package streamtools

import (
	"github.com/bitly/go-simplejson"
	"log"
)

// Block is the basic interface for processing units in streamtools
type Block interface {
	// blockRoutine is the central processing routine for a block. All the work gets done in here
	blockRoutine()
	// a set of accessors are provided so that a block creator can access certain aspects of a block
	getID() string
	getBlockType() string
	getOutChan() chan *simplejson.Json
	getInChan() chan *simplejson.Json
	getRouteChan(string) chan routeResponse
	getRoutes() []string
	// some aspects of a block can also be set by the block creator
	setOutChan(chan *simplejson.Json)
	setInChan(chan *simplejson.Json)
	setID(string)
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
	outChan chan *simplejson.Json
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

func (self *AbstractBlock) getOutChan() chan *simplejson.Json {
	return self.outChan
}

func (self *AbstractBlock) setInChan(inChan chan *simplejson.Json) {
	self.inChan = inChan
}

func (self *AbstractBlock) setOutChan(outChan chan *simplejson.Json) {
	self.outChan = outChan
}

func (self *AbstractBlock) setID(id string) {
	self.ID = id
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

// NewBlock returns a block from the block factory specified in the library
func NewBlock(blockType string) Block {
	blockTemplate, ok := library[blockType]
	if !ok {
		log.Fatal("couldn't find block", blockType)
	}
	return blockTemplate.blockFactory()
}
