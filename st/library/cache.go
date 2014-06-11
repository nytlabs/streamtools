package library

import (
	"container/heap"
	"encoding/json"
	"errors"
	"time"

	"github.com/nytlabs/gojee"                 // jee
	"github.com/nytlabs/streamtools/st/blocks" // blocks
	"github.com/nytlabs/streamtools/st/util"   // util
)

// specify those channels we're going to use to communicate with streamtools
type Cache struct {
	blocks.Block
	querylookup chan blocks.Query
	queryrule   chan blocks.MsgChan
	inrule      blocks.MsgChan
	in          blocks.MsgChan
	lookup      blocks.MsgChan
	keys        chan blocks.MsgChan
	values      chan blocks.MsgChan
	dump        chan blocks.MsgChan
	out         blocks.MsgChan
	quit        blocks.MsgChan
}

// we need to build a simple factory so that streamtools can make new blocks of this kind
func NewCache() blocks.BlockInterface {
	return &Cache{}
}

type item struct {
	value    interface{}
	lastSeen time.Time
}

func (i item) MarshalJSON() ([]byte, error) {
	return json.Marshal(i.value)
}

// Cacheup is called once before running the block. We build up the channels and specify what kind of block this is.
func (b *Cache) Setup() {
	b.Kind = "Core"
	b.Desc = "stores a set of dictionary values queryable on key"
	b.inrule = b.InRoute("rule")
	b.queryrule = b.QueryRoute("rule")
	b.querylookup = b.QueryParamRoute("lookup")
	b.quit = b.Quit()
	b.out = b.Broadcast()

	b.in = b.InRoute("in")
	b.lookup = b.InRoute("lookup")
	b.keys = b.QueryRoute("keys")
	b.values = b.QueryRoute("values")
	b.dump = b.QueryRoute("dump")
}

func extractAndUpdate(k string, values map[string]item, ttlQueue *PriorityQueue) (map[string]interface{}, error) {
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
	out := map[string]interface{}{
		"key":   k,
		"value": v,
	}
	return out, nil

}

// Run is the block's main loop. Here we listen on the different channels we set up.
func (b *Cache) Run() {
	var keyPath, valuePath, ttlString string
	var ttl time.Duration
	cache := make(map[string]item)
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
				continue
			}
			k, ok := kI.(string)
			if !ok {
				b.Error(err)
				continue
			}
			out, err := extractAndUpdate(k, cache, ttlQueue)
			if err != nil {
				b.Error(err)
				continue
			}
			b.out <- out
		case q := <-b.querylookup:
			k, ok := q.Params["key"]
			if !ok {
				b.Error(errors.New("Must specify a key to lookup"))
			}
			for _, ki := range k {
				out, err := extractAndUpdate(ki, cache, ttlQueue)
				if err != nil {
					b.Error(err)
					continue
				}
				q.RespChan <- out
			}

		case responseChan := <-b.keys:
			keys := make([]string, len(cache))
			i := 0
			for key := range cache {
				keys[i] = key
				i++
			}

			responseChan <- map[string]interface{}{
				"keys": keys,
			}

		case responseChan := <-b.values:
			values := make([]interface{}, len(cache))
			i := 0
			for _, item := range cache {
				values[i] = item
				i++
			}
			responseChan <- map[string]interface{}{
				"values": values,
			}

		case responseChan := <-b.dump:
			responseChan <- map[string]interface{}{
				"dump": cache,
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
			cache[k] = item{
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
			if i.t.Equal(cache[k].lastSeen) {
				delete(cache, k)
			}
		}
	}
}
