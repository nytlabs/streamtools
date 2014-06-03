package library

import (
	"container/heap"
	"errors"
	"time"

	"github.com/nytlabs/gojee"                 // jee
	"github.com/nytlabs/streamtools/st/blocks" // blocks
	"github.com/nytlabs/streamtools/st/util"   // util
)

// specify those channels we're going to use to communicate with streamtools
type Histogram struct {
	blocks.Block
	queryrule chan blocks.MsgChan
	historule chan blocks.MsgChan
	inrule    blocks.MsgChan
	inpoll    blocks.MsgChan
	in        blocks.MsgChan
	out       blocks.MsgChan
	quit      blocks.MsgChan
}

// we need to build a simple factory so that streamtools can make new blocks of this kind
func NewHistogram() blocks.BlockInterface {
	return &Histogram{}
}

func buildHistogram(histogram map[string]*PriorityQueue) interface{} {
	var data interface{}
	var buckets []interface{}
	buckets = make([]interface{}, len(histogram))

	i := 0
	for k, pq := range histogram {
		var bucket interface{}
		bucket = map[string]interface{}{
			"Count": float64(len(*pq)),
			"Label": k,
		}
		buckets[i] = bucket
		i++
	}

	data = map[string]interface{}{
		"Histogram": buckets,
	}
	return data
}

// Setup is called once before running the block. We build up the channels and specify what kind of block this is.
func (b *Histogram) Setup() {
	b.Kind = "Stats"
	b.Desc = "builds a non-stationary histogram of inbound messages for a specified path"
	b.in = b.InRoute("in")
	b.inrule = b.InRoute("rule")
	b.queryrule = b.QueryRoute("rule")
	b.historule = b.QueryRoute("histogram")
	b.inpoll = b.InRoute("poll")
	b.quit = b.Quit()
	b.out = b.Broadcast()
}

// Run is the block's main loop. Here we listen on the different channels we set up.
func (b *Histogram) Run() {
	var tree *jee.TokenTree
	var path string
	waitTimer := time.NewTimer(100 * time.Millisecond)
	window := time.Duration(0)

	histogram := map[string]*PriorityQueue{}
	emptyByte := make([]byte, 0)
	for {
		select {
		case ruleI := <-b.inrule:
			// window
			windowString, err := util.ParseString(ruleI, "Window")
			if err != nil {
				b.Error(err)
			}
			window, err = time.ParseDuration(windowString)
			if err != nil {
				b.Error(err)
			}
			path, err = util.ParseString(ruleI, "Path")
			tree, err = util.BuildTokenTree(path)
			if err != nil {
				b.Error(err)
				break
			}

		case <-b.quit:
			// quit the block
			return
		case msg := <-b.in:
			if tree == nil {
				continue
			}
			v, err := jee.Eval(tree, msg)
			if err != nil {
				b.Error(err)
				break
			}
			valueString, ok := v.(string)
			if !ok {
				b.Error(errors.New("nil value against" + path + " - ignoring"))
				break
			}

			if pq, ok := histogram[valueString]; ok {
				queueMessage := &PQMessage{
					val: &emptyByte,
					t:   time.Now(),
				}
				heap.Push(pq, queueMessage)
			} else {
				pq := &PriorityQueue{}
				heap.Init(pq)
				histogram[valueString] = pq
				queueMessage := &PQMessage{
					val: &emptyByte,
					t:   time.Now(),
				}
				heap.Push(pq, queueMessage)
			}
		case <-waitTimer.C:

		case <-b.inpoll:
			// deal with a poll request
			data := buildHistogram(histogram)
			b.out <- data
		case MsgChan := <-b.queryrule:
			// deal with a query request
			out := map[string]interface{}{
				"Window": window.String(),
				"Path":   path,
			}
			MsgChan <- out
		case MsgChan := <-b.historule:
			data := buildHistogram(histogram)
			MsgChan <- data
		}
		for _, pq := range histogram {
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
}
