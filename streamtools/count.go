package streamtools

import (
	"container/heap"
	"github.com/bitly/go-simplejson"
	"log"
	"time"
)

func Count(inChan chan *simplejson.Json, ruleChan chan *simplejson.Json, queryChan chan stateQuery) {
	// block until we recieve a rule
	params = <-ruleChan
	windowSeconds, err := params.Get("window").Float64()
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
		case params = <-ruleChan:
		case query = <-queryChan:
			out, err := simplejson.NewJson([]byte("{}"))
			if err != nil {
				log.Fatal(err.Error())
			}
			out.Set("count", len(*pq))
			query.responseChan <- out
		case <-inChan:
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
				waitTimer.Reset(diff)
				break
			}
		}
	}
}
