package blocks

import (
	"container/heap"
	"log"
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

	data := &histogramData{}

	var rule *histogramRule

	waitTimer := time.NewTimer(100 * time.Millisecond)
	window := time.Duration(0)

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
		case msg := <-b.Routes["get_rule"]:
			if rule == nil {
				marshal(msg, &histogramRule{})
			} else {
				marshal(msg, rule)
			}
		case ruleUpdate := <-b.Routes["set_rule"]:
			if rule == nil {
				rule = &histogramRule{}
			}

			unmarshal(ruleUpdate, rule)
			window = time.Duration(rule.Window) * time.Second
		case msg := <-b.InChan:
			if rule == nil {
				break
			}

			valueSlice := getKeyValues(msg, rule.Key)
			// we need to guard against the possibility that the key is not in
			// the message
			if len(valueSlice) == 0 {
				log.Println("could not find", rule.Key, "in message")
				break
			}

			value := valueSlice[0]
			valueString, ok := value.(string)
			if !ok {
				log.Println("nil value against", rule.Key, " - ignoring")
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
		case <-b.QuitChan:
			quit(b)
			return
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
