package streamtools

import (
	"github.com/bitly/go-simplejson"
	"log"
)

type Block interface {
	blockRoutine()
	getID() string
}

type AbstractBlock struct {
	name      string
	ID        string
	blockType string
	inChan    chan *simplejson.Json
	outChan   chan *simplejson.Json
	ruleChan  chan *simplejson.Json
	queryChan chan StreamtoolsQuery
}

func (self *AbstractBlock) getID() string {
	return self.ID
}

func NewBlock(blockType string) Block {
	blockTemplate, ok := library[blockType]
	if !ok {
		log.Fatal("couldn't find block", blockType)
	}
	return blockTemplate.blockFactory()
}
