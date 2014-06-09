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
type Sync struct {
	blocks.Block
	queryrule chan blocks.MsgChan
	inrule    blocks.MsgChan
	in        blocks.MsgChan
	out       blocks.MsgChan
	quit      blocks.MsgChan
}

// we need to build a simple factory so that streamtools can make new blocks of this kind
func NewSync() blocks.BlockInterface {
	return &Sync{}
}

// Setup is called once before running the block. We build up the channels and specify what kind of block this is.
func (b *Sync) Setup() {
	b.Kind = "Core"
	b.Desc = "takes an disordered stream and creates a properly timed, ordered stream at the expense of introducing a lag"
	b.in = b.InRoute("in")
	b.inrule = b.InRoute("rule")
	b.queryrule = b.QueryRoute("rule")
	b.quit = b.Quit()
	b.out = b.Broadcast()
}

// Run is the block's main loop. Here we listen on the different channels we set up.
func (b *Sync) Run() {
	var lagString, path string
	var tree *jee.TokenTree
	lag := time.Duration(0)
	emitTick := time.NewTimer(500 * time.Millisecond)
	pq := &PriorityQueue{}
	heap.Init(pq)
	for {
		select {
		case <-emitTick.C:
		case ruleI := <-b.inrule:
			// set a parameter of the block
			lagString, err := util.ParseString(ruleI, "Lag")
			if err != nil {
				b.Error(err)
				break
			}
			lag, err = time.ParseDuration(lagString)
			if err != nil {
				b.Error(err)
				continue
			}
			path, err = util.ParseString(ruleI, "Path")
			if err != nil {
				b.Error(err)
				break
			}
			// build the parser for the model
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
		case <-b.quit:
			// quit the block
			return
		case msg := <-b.in:
			// deal with inbound data
			if tree == nil {
				break
			}
			tI, err := jee.Eval(tree, interface{}(msg))
			if err != nil {
				b.Error(err)
			}
			t, ok := tI.(float64)
			if !ok {
				b.Error(errors.New("couldn't convert time value to float64"))
				continue
			}
			ms := time.Unix(0, int64(t*1000000))
			queueMessage := &PQMessage{
				val: msg,
				t:   ms,
			}
			heap.Push(pq, queueMessage)

		case MsgChan := <-b.queryrule:
			// deal with a query request
			MsgChan <- map[string]interface{}{
				"Lag":  lagString,
				"Path": path,
			}

		}
		now := time.Now()
		for {
			item, diff := pq.PeekAndShift(now, lag)
			if item == nil {
				// then the queue is empty. Pause for 5 seconds before checking again
				if diff == 0 {
					diff = time.Duration(500) * time.Millisecond
				}
				emitTick.Reset(diff)
				break
			}
			b.out <- item.(*PQMessage).val.(map[string]interface{})
		}

	}
}
