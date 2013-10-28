package blocks

import (
	"container/heap"
	"time"
)

func Histogram(b *Block) {

	type histogramRule struct {
		Window int
		Key    string
	}

	type histogramBucket struct {
		Count int
		Label string
	}

	type histogramData struct {
		Histogram []histogramBucket
	}

	rule := &histogramRule{}
	data := &histogramData{}

	// block until we recieve a rule
	unmarshal(<-b.Routes["set_rule"], &rule)

	window := time.Duration(rule.Window) * time.Second
	waitTimer := time.NewTimer(100 * time.Millisecond)

	histogram := map[string]*PriorityQueue{}
	emptyByte := make([]byte, 0)

	for {
		select {
		case query := <-b.Routes["histogram"]:
			data.Histogram = make([]histogramBucket, len(histogram))
			i := 0
			for k, pq := range histogram {
				bucket := histogramBucket{
					Count: len(*pq),
					Label: k,
				}
				data.Histogram[i] = bucket
				i++
			}
			marshal(query, data)
		case ruleUpdate := <-b.Routes["set_rule"]:
			unmarshal(ruleUpdate, &rule)
			window = time.Duration(rule.Window) * time.Second
		case msg := <-b.InChan:
			value := getKeyValues(msg.Interface(), rule.Key)[0]
			valueString := value.(string)

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
