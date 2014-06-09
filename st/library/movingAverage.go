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
type MovingAverage struct {
	blocks.Block
	queryrule chan blocks.MsgChan
	queryavg  chan blocks.MsgChan
	inrule    blocks.MsgChan
	inpoll    blocks.MsgChan
	in        blocks.MsgChan
	out       blocks.MsgChan
	quit      blocks.MsgChan
}

// we need to build a simple factory so that streamtools can make new blocks of this kind
func NewMovingAverage() blocks.BlockInterface {
	return &MovingAverage{}
}

// Setup is called once before running the block. We build up the channels and specify what kind of block this is.
func (b *MovingAverage) Setup() {
	b.Kind = "Stats"
	b.Desc = "performs a moving average of the values specified by the Path over the duration of the Window"
	b.in = b.InRoute("in")
	b.inrule = b.InRoute("rule")
	b.queryrule = b.QueryRoute("rule")
	b.queryavg = b.QueryRoute("average")
	b.inpoll = b.InRoute("poll")
	b.quit = b.Quit()
	b.out = b.Broadcast()
}

func pqAverage(pq *PriorityQueue) float64 {
	var sum float64
	sum = 0
	for _, pqmsg := range *pq {
		v := pqmsg.val
		val, ok := v.(float64)
		if !ok {
			continue
		}
		sum += val
	}
	N := float64(len(*pq))
	return sum / N
}

// Run is the block's main loop. Here we listen on the different channels we set up.
func (b *MovingAverage) Run() {
	var tree *jee.TokenTree
	var path, windowString string
	var err error
	window := time.Duration(0)
	waitTimer := time.NewTimer(100 * time.Millisecond)

	pq := &PriorityQueue{}
	heap.Init(pq)

	for {
		select {
		case ruleI := <-b.inrule:
			// set a parameter of the block
			path, err = util.ParseString(ruleI, "Path")
			if err != nil {
				b.Error(err)
			}
			tree, err = util.BuildTokenTree(path)
			if err != nil {
				b.Error(err)
				break
			}
			windowString, err = util.ParseString(ruleI, "Window")
			if err != nil {
				b.Error(err)
			}
			window, err = time.ParseDuration(windowString)
			if err != nil {
				b.Error(err)
			}
		case <-b.quit:
			// quit the block
			return
		case msg := <-b.in:
			// deal with inbound data
			if tree == nil {
				break
			}
			val, err := jee.Eval(tree, msg)
			if err != nil {
				b.Error(err)
				break
			}
			// TODO make this a type swtich and convert anything we can to a
			// float
			val, ok := val.(float64)
			if !ok {
				b.Error(errors.New("trying to put a non-float into the moving average"))
				continue
			}
			queueMessage := &PQMessage{
				val: val,
				t:   time.Now(),
			}
			heap.Push(pq, queueMessage)
		case <-b.inpoll:
			// deal with a poll request
			outMsg := map[string]interface{}{
				"Average": pqAverage(pq),
			}
			b.out <- outMsg
		case c := <-b.queryavg:
			outMsg := map[string]interface{}{
				"Average": pqAverage(pq),
			}
			c <- outMsg
		case c := <-b.queryrule:
			// deal with a query request
			c <- map[string]interface{}{
				"Path":   path,
				"Window": windowString,
			}
		case <-waitTimer.C:
		}
		for {
			pqMsg, diff := pq.PeekAndShift(time.Now(), window)
			if pqMsg == nil {
				// either the queue is empty, or it's not time to emit
				if diff == 0 {
					// then the queue is empty. Pause for 5 seconds before checking again
					diff = time.Duration(500) * time.Millisecond
				}
				waitTimer.Reset(diff)
				break
			}
		}
	}
}
