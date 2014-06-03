package library

import (
	"container/heap"
	"errors"
	"time"

	"github.com/nytlabs/gojee"
	"github.com/nytlabs/streamtools/st/blocks" // blocks
	"github.com/nytlabs/streamtools/st/util"
)

// specify those channels we're going to use to communicate with streamtools
type PackByValue struct {
	blocks.Block
	queryrule chan blocks.MsgChan
	inrule    blocks.MsgChan
	in        blocks.MsgChan
	out       blocks.MsgChan
	quit      blocks.MsgChan
}

// we need to build a simple factory so that streamtools can make new blocks of this kind
func NewPackByValue() blocks.BlockInterface {
	return &PackByValue{}
}

// Setup is called once before running the block. We build up the channels and specify what kind of block this is.
func (b *PackByValue) Setup() {
	b.Kind = "Core"
	b.Desc = "groups messages together based on a common value, similar to 'group-by' in other languages"
	b.in = b.InRoute("in")
	b.inrule = b.InRoute("rule")
	b.queryrule = b.QueryRoute("rule")
	b.quit = b.Quit()
	b.out = b.Broadcast()
}

// Run is the block's main loop. Here we listen on the different channels we set up.
func (b *PackByValue) Run() {
	var tree *jee.TokenTree
	var emitAfter, path string
	var err error

	afterDuration := time.Duration(0)
	waitTimer := time.NewTimer(100 * time.Millisecond)
	bunches := make(map[string][]interface{})
	pq := &PriorityQueue{}
	heap.Init(pq)

	for {
		select {
		case <-waitTimer.C:
		case ruleI := <-b.inrule:
			// set a parameter of the block
			rule, ok := ruleI.(map[string]interface{})
			if !ok {
				b.Error(errors.New("coudln't assert rule to map"))
				continue
			}
			path, err = util.ParseString(rule, "Path")
			if err != nil {
				b.Error(err)
				continue
			}
			token, err := jee.Lexer(path)
			if err != nil {
				b.Error(err)
				continue
			}
			tree, err = jee.Parser(token)
			if err != nil {
				b.Error(err)
				continue
			}
			emitAfter, err = util.ParseString(rule, "EmitAfter")
			if err != nil {
				b.Error(err)
				continue
			}
			afterDuration, err = time.ParseDuration(emitAfter)
			if err != nil {
				b.Error(err)
				continue
			}

		case <-b.quit:
			// quit the block
			return
		case msg := <-b.in:
			// deal with inbound data
			if tree == nil {
				break
			}

			id, err := jee.Eval(tree, msg)
			if err != nil {
				b.Error(err)
				continue
			}
			idStr, ok := id.(string)
			if !ok {
				b.Error(errors.New("could not assert id to string"))
				break
			}
			if len(bunches[idStr]) > 0 {
				bunches[idStr] = append(bunches[idStr], msg)
			} else {
				bunches[idStr] = []interface{}{msg}
			}

			val := map[string]interface{}{
				"id":     idStr,
				"length": len(bunches[idStr]),
			}

			queueMessage := &PQMessage{
				val: val,
				t:   time.Now(),
			}
			heap.Push(pq, queueMessage)
		case c := <-b.queryrule:
			// deal with a query request
			c <- map[string]interface{}{
				"Path":      path,
				"EmitAfter": emitAfter,
			}
		}
		for {
			pqMsg, diff := pq.PeekAndShift(time.Now(), afterDuration)
			if pqMsg == nil {
				// either the queue is empty, or it's not time to emit
				waitTimer.Reset(diff)
				break
			}
			v := pqMsg.(*PQMessage).val.(map[string]interface{})

			l, ok := v["length"]
			if !ok {
				b.Error(errors.New("couldn't find length in message"))
				continue
			}
			lInt := l.(int)
			id, ok := v["id"]
			if !ok {
				b.Error(errors.New("couldn't find id in message"))
				continue
			}
			idStr := id.(string)
			if lInt == len(bunches[idStr]) {
				// we've not seen anything since putting this message in the queue

				msg := map[string]interface{}{
					"Pack": bunches[idStr],
				}

				b.out <- msg

				delete(bunches, idStr)
			}
		}
	}
}
