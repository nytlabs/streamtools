package streamtools

import (
	"github.com/bitly/go-simplejson"
	"log"
)

type Block interface {
	blockRoutine()
	getID() string
	getBlockType() string
	getOutChan() chan *simplejson.Json
	getInChan() chan *simplejson.Json
	getQueryChan() chan query
	setOutChan(chan *simplejson.Json)
	setInChan(chan *simplejson.Json)
}

type AbstractBlock struct {
	name      string
	ID        string
	blockType string
	inChan    chan *simplejson.Json
	outChan   chan *simplejson.Json
	ruleChan  chan *simplejson.Json
	queryChan chan query
}

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

func (self *AbstractBlock) getQueryChan() chan query {
	return self.queryChan
}

func (self *AbstractBlock) setInChan(inChan chan *simplejson.Json) {
	self.inChan = inChan
}

func (self *AbstractBlock) setOutChan(outChan chan *simplejson.Json) {
	self.outChan = outChan
}

func NewBlock(blockType string) Block {
	blockTemplate, ok := library[blockType]
	if !ok {
		log.Fatal("couldn't find block", blockType)
	}
	return blockTemplate.blockFactory()
}
