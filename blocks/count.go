package blocks

import (
	"container/heap"
	"github.com/bitly/go-simplejson"
	"log"
	"time"
)

type CountBlock struct {
	AbstractBlock
}

func (b CountBlock) BlockRoutine() {
	log.Println("starting count block")

	query := <-b.routes["getRules"]
	params := query.Msg
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
		case query = <-b.routes["count"]:
			out, err := simplejson.NewJson([]byte("{}"))
			if err != nil {
				log.Fatal(err.Error())
			}
			out.Set("count", len(*pq))
			query.ResponseChan <- out
		case <-b.inChan:
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

func NewCount() Block {
	b := new(CountBlock)
	b.blockType = "count"
	b.inChan = make(chan *simplejson.Json)
	b.routes = map[string]chan RouteResponse{
		"setRule":  make(chan RouteResponse),
		"getRules": make(chan RouteResponse),
		"count":    make(chan RouteResponse),
	}
	return b
}
