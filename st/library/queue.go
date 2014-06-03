package library

import (
	"container/heap"
	"time"

	"github.com/nytlabs/streamtools/st/blocks" // blocks
)

// specify those channels we're going to use to communicate with streamtools
type Queue struct {
	blocks.Block
	queryPop  chan blocks.MsgChan
	queryPeek chan blocks.MsgChan
	inPush    blocks.MsgChan
	inPop     blocks.MsgChan
	out       blocks.MsgChan
	quit      blocks.MsgChan
}

// we need to build a simple factory so that streamtools can make new blocks of this kind
func NewQueue() blocks.BlockInterface {
	return &Queue{}
}

// Setup is called once before running the block. We build up the channels and specify what kind of block this is.
func (b *Queue) Setup() {
	b.Kind = "Queues"
	b.Desc = "FIFO queue allowing push & pop on streams plus popping from a query"
	b.inPush = b.InRoute("push")
	b.inPop = b.InRoute("pop")
	b.queryPop = b.QueryRoute("pop")
	b.queryPeek = b.QueryRoute("peek")
	b.quit = b.Quit()
	b.out = b.Broadcast()
}

// Run is the block's main loop. Here we listen on the different channels we set up.
func (b *Queue) Run() {
	pq := &PriorityQueue{}
	heap.Init(pq)
	for {
		select {
		case <-b.quit:
			// quit the block
			return
		case msg := <-b.inPush:
			queueMessage := &PQMessage{
				val: msg,
				t:   time.Now(),
			}
			heap.Push(pq, queueMessage)
		case <-b.inPop:
			if len(*pq) == 0 {
				continue
			}
			msg := heap.Pop(pq).(*PQMessage).val
			b.out <- msg
		case MsgChan := <-b.queryPop:
			var msg interface{}
			if len(*pq) > 0 {
				msg = heap.Pop(pq).(*PQMessage).val
			}
			MsgChan <- msg
		case MsgChan := <-b.queryPeek:
			var msg interface{}
			if len(*pq) > 0 {
				msg = pq.Peek().(*PQMessage).val
			}
			MsgChan <- msg
		}
	}
}
