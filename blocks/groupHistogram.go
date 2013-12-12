package blocks

import (
	"container/heap"
	"log"
	"time"
)

// GroupHistogram is a group of histograms, where each histogram is indexed by
// one field in the incoming message, and the histogram captures a distrbuition
// over another field.
func GroupHistogram(b *Block) {

	// rule to setup the block
	type groupHistogramRule struct {
		Window   int
		GroupKey string
		Key      string
	}

	// the form of the histogram query
	type histogramQuery struct {
		GroupKey string
	}

	type histogramBucket struct {
		Count int
		Label string
	}

	// the form of the returned histogram in response to the histogram query
	type groupHistogramData struct {
		Histogram []histogramBucket
		GroupKey  string
	}

	type histogram map[string]*PriorityQueue

	type groupHistogram map[string]histogram

	data := &groupHistogramData{}

	var rule *groupHistogramRule

	waitTimer := time.NewTimer(100 * time.Millisecond)
	window := time.Duration(0)

	emptyByte := make([]byte, 0)

	histograms := make(map[string]histogram)

	for {
		select {
		case msg := <-b.Routes["histogram"]:
			query, ok := msg.(RouteResponse)
			if !ok {
				break
			}

			var q histogramQuery
			decode(query.Msg, &q)
			h, ok := histograms[q.GroupKey]
			if !ok {
				log.Println("could not find requested histogram in group",
					q.GroupKey)
				break
			}
			data.Histogram = make([]histogramBucket, len(h))
			i := 0
			for k, pq := range h {
				bucket := histogramBucket{
					Count: len(*pq),
					Label: k,
				}
				data.Histogram[i] = bucket
				i++
			}
			data.GroupKey = q.GroupKey
			marshal(query, data)
		case query := <-b.Routes["list"]:
			var keys []string
			for k := range histograms {
				keys = append(keys, k)
			}
			marshal(query, keys)
		case msg := <-b.Routes["get_rule"]:
			if rule == nil {
				marshal(msg, &groupHistogramRule{})
			} else {
				marshal(msg, rule)
			}
		case ruleUpdate := <-b.Routes["set_rule"]:
			if rule == nil {
				rule = &groupHistogramRule{}
			}

			unmarshal(ruleUpdate, rule)
			window = time.Duration(rule.Window) * time.Second
		case msg := <-b.InChan:
			if rule == nil {
				break
			}

			groupKey := getKeyValues(msg, rule.GroupKey)[0]
			groupKeyString, ok := groupKey.(string)
			if !ok {
				log.Println("groupKey must be a string")
				break
			}
			value := getKeyValues(msg, rule.Key)[0]
			valueString, ok := value.(string)
			if !ok {
				log.Println("nil value against", rule.Key, " - ignoring")
				break
			}

			var h histogram
			h, ok = histograms[groupKeyString]
			if !ok {
				log.Println("making new histogram with groupKey",
					groupKeyString)
				h = make(map[string]*PriorityQueue)
				histograms[groupKeyString] = h
			}

			if pq, ok := h[valueString]; ok {
				queueMessage := &PQMessage{
					val: &emptyByte,
					t:   time.Now(),
				}
				heap.Push(pq, queueMessage)
			} else {
				log.Println("making new value in group", groupKeyString, "with value", valueString)
				pq := &PriorityQueue{}
				heap.Init(pq)
				h[valueString] = pq
				queueMessage := &PQMessage{
					val: &emptyByte,
					t:   time.Now(),
				}
				heap.Push(pq, queueMessage)
			}
		case <-waitTimer.C:
		}
		for _, h := range histograms {
			for _, pq := range h {
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
}
