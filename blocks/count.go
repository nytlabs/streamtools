package blocks

import (
	"container/heap"
	"github.com/bitly/go-simplejson"
	"log"
	"time"
)

func Count(b *Block) {
	// block until we recieve a rule
	params := <-b.Routes["params"]
	windowSeconds, err := params.Msg.Get("window").Float64()
	params.ResponseChan <- nil
	if err != nil {
		log.Fatal(err.Error())
	}
	window := time.Duration(windowSeconds) * time.Second
	waitTimer := time.NewTimer(100 * time.Millisecond)

	pq := &PriorityQueue{}
	heap.Init(pq)

	emptyByte := make([]byte, 0)

	for {
		select {
		case query := <-b.Routes["count"]:
			out, err := simplejson.NewJson([]byte("{}"))
			if err != nil {
				log.Fatal(err.Error())
			}
			out.Set("count", len(*pq))
			query.ResponseChan <- out
		case <-b.InChan:
			queueMessage := &PQMessage{
				val: &emptyByte,
				t:   time.Now(),
			}
			if err != nil {
				log.Fatal(err.Error())
			}
			heap.Push(pq, queueMessage)
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
