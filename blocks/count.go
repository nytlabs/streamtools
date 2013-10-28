package blocks

import (
	"container/heap"
	"time"
)

func Count(b *Block) {

	type countRule struct {
		Window int
	}

	type countData struct {
		Count int
	}

	rule := &countRule{}
	data := &countData{Count: 0}

	// block until we recieve a rule
	unmarshal(<-b.Routes["set_rule"], &rule)

	window := time.Duration(rule.Window) * time.Second
	waitTimer := time.NewTimer(100 * time.Millisecond)

	pq := &PriorityQueue{}
	heap.Init(pq)

	emptyByte := make([]byte, 0)

	for {
		select {
		case query := <-b.Routes["count"]:
			data.Count = len(*pq)
			marshal(query, data)
		case ruleUpdate := <-b.Routes["set_rule"]:
			unmarshal(ruleUpdate, &rule)
			window = time.Duration(rule.Window) * time.Second
		case <-b.QuitChan:
			quit(b)
			return
		case <-b.InChan:
			queueMessage := &PQMessage{
				val: &emptyByte,
				t:   time.Now(),
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
