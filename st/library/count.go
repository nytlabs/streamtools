package library

import (
	"container/heap"
	"github.com/nytlabs/streamtools/st/blocks" // blocks
	"time"
)

type Count struct {
	blocks.Block
	queryrule chan chan interface{}
	inrule    chan interface{}
	inpoll    chan interface{}
	in        chan interface{}
	out       chan interface{}
	quit      chan interface{}
	windowStr string
}

// a bit of boilerplate for streamtools
func NewCount() blocks.BlockInterface {
	return &Count{}
}

func (b *Count) Setup() {
	b.Kind = "Count"
	b.in = b.InRoute("in")
	b.inrule = b.InRoute("rule")
	b.inpoll = b.InRoute("poll")
	b.queryrule = b.QueryRoute("rule")
	b.quit = b.Quit()
	b.out = b.Broadcast()
}

func (b *Count) Run() {
	waitTimer := time.NewTimer(100 * time.Millisecond)
	pq := &PriorityQueue{}
	heap.Init(pq)
	window := time.Duration(0)

	for {
		select {
		case <-waitTimer.C:
		case rule := <-b.inrule:
			b.windowStr = rule.(map[string]string)["Window"]
			window, _ = time.ParseDuration(b.windowStr)
		case <-b.quit:
			return
		case <-b.in:
			empty := make([]byte, 0)
			queueMessage := &PQMessage{
				val: &empty,
				t:   time.Now(),
			}
			heap.Push(pq, queueMessage)
		case <-b.inpoll:
			b.out <- map[string]interface{}{
				"Count": len(*pq),
			}
		case c := <-b.queryrule:
			c <- map[string]string{
				"Window": b.windowStr,
			}
		}
		for {
			pqMsg, diff := pq.PeekAndShift(time.Now(), window)
			if pqMsg == nil {
				// either the queue is empty, or it"s not time to emit
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
