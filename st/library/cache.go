package library

import (
	"container/heap"
	"errors"
	"github.com/nytlabs/gojee"                 // jee
	"github.com/nytlabs/streamtools/st/blocks" // blocks
	"github.com/nytlabs/streamtools/st/util"   // util
	"time"
)

// specify those channels we're going to use to communicate with streamtools
type Cache struct {
	blocks.Block
	queryrule chan chan interface{}
	inrule    chan interface{}
	in        chan interface{}
	lookup    chan interface{}
	out       chan interface{}
	quit      chan interface{}
}

// we need to build a simple factory so that streamtools can make new blocks of this kind
func NewCache() blocks.BlockInterface {
	return &Cache{}
}

type item struct {
	value    interface{}
	lastSeen time.Time
}

// Cacheup is called once before running the block. We build up the channels and specify what kind of block this is.
func (b *Cache) Setup() {
	b.Kind = "Cache"
	b.Desc = "stores a set of dictionary values queryable on key"
	b.inrule = b.InRoute("rule")
	b.queryrule = b.QueryRoute("rule")
	b.quit = b.Quit()
	b.out = b.Broadcast()

	b.in = b.InRoute("in")
	b.lookup = b.InRoute("lookup")
}

// Run is the block's main loop. Here we listen on the different channels we set up.
func (b *Cache) Run() {
	var keyPath, valuePath, ttlString string
	var ttl time.Duration
	values := make(map[string]item)
	ttlQueue := &PriorityQueue{}

	var keyTree, valueTree *jee.TokenTree
	var err error
	emitTick := time.NewTimer(500 * time.Millisecond)
	for {
		select {
		case <-emitTick.C:

		case ruleI := <-b.inrule:
			keyPath, err = util.ParseString(ruleI, "KeyPath")
			keyTree, err = util.BuildTokenTree(keyPath)
			if err != nil {
				b.Error(err)
				break
			}
			valuePath, err = util.ParseString(ruleI, "ValuePath")
			valueTree, err = util.BuildTokenTree(valuePath)
			if err != nil {
				b.Error(err)
				break
			}
			ttlString, err = util.ParseString(ruleI, "TimeToLive")
			if err != nil {
				b.Error(err)
				break
			}
			ttl, err = time.ParseDuration(ttlString)
			if err != nil {
				b.Error(err)
				break
			}
		case <-b.quit:
			return
		case msg := <-b.lookup:
			if keyTree == nil {
				continue
			}
			kI, err := jee.Eval(keyTree, msg)
			if err != nil {
				b.Error(err)
				break
			}
			k, ok := kI.(string)
			if !ok {
				b.Error(errors.New("key must be a string"))
				continue
			}
			i, ok := values[k]
			var v interface{}
			if ok {
				v = i.value
				now := time.Now()
				i.lastSeen = now
				queueMessage := &PQMessage{
					val: k,
					t:   now,
				}
				heap.Push(ttlQueue, queueMessage)
			}
			b.out <- map[string]interface{}{
				"key":   k,
				"value": v,
			}
		case msg := <-b.in:
			if keyTree == nil {
				continue
			}
			if valueTree == nil {
				continue
			}
			kI, err := jee.Eval(keyTree, msg)
			if err != nil {
				b.Error(err)
				break
			}
			k, ok := kI.(string)
			if !ok {
				b.Error(errors.New("key must be a string"))
				continue
			}
			v, err := jee.Eval(valueTree, msg)
			if err != nil {
				b.Error(err)
				break
			}
			now := time.Now()
			values[k] = item{
				value:    v,
				lastSeen: now,
			}
			queueMessage := &PQMessage{
				val: k,
				t:   now,
			}
			heap.Push(ttlQueue, queueMessage)
		case responseChan := <-b.queryrule:
			// deal with a query request
			responseChan <- map[string]interface{}{
				"KeyPath":    keyPath,
				"ValuePath":  valuePath,
				"TimeToLive": ttlString,
			}
		}
		now := time.Now()
		for {
			itemI, diff := ttlQueue.PeekAndShift(now, ttl)
			if itemI == nil {
				// then the queue is empty. don't check again for 5s
				if diff == 0 {
					diff = time.Duration(500) * time.Millisecond
				}
				emitTick.Reset(diff)
				break
			}
			i := itemI.(*PQMessage)
			k := i.val.(string)
			if i.t.Equal(values[k].lastSeen) {
				delete(values, k)
			}
		}
	}
}
