package blocks

import (
	"container/heap"
	"log"
	"time"
)

// Count uses a priority queue to count the number of messages that have been sent
// to the count block over a duration of time in seconds.
//
// Note that this is an exact count and therefore has O(N) memory requirements.
func Count(b *Block) {

	var err error

	type countRule struct {
		Window string
	}

	type countData struct {
		Count int
	}

	data := &countData{Count: 0}
	var rule *countRule

	window := time.Duration(0)
	waitTimer := time.NewTimer(100 * time.Millisecond)

	pq := &PriorityQueue{}
	heap.Init(pq)

	emptyByte := make([]byte, 0)

	for {
		select {
		case query := <-b.Routes["count"]:
			data.Count = len(*pq)
			marshal(query, data)
		case <-b.Routes["poll"]:
			outMsg := map[string]interface{}{
				"Count": len(*pq),
			}
			out := BMsg{
				Msg: outMsg,
			}
			broadcast(b.OutChans, out)
		case ruleUpdate := <-b.Routes["set_rule"]:
			if rule == nil {
				rule = &countRule{}
			}
			unmarshal(ruleUpdate, rule)
			window, err = time.ParseDuration(rule.Window)
			if err != nil {
				log.Println(err.Error())
			}
		case msg := <-b.Routes["get_rule"]:
			if rule == nil {
				marshal(msg, &countRule{})
			} else {
				marshal(msg, rule)
			}
		case <-b.QuitChan:
			quit(b)
			return
		case msg := <-b.AddChan:
			updateOutChans(msg, b)
		case <-b.InChan:
			if rule == nil {
				break
			}

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
