package blocks

import (
	"container/heap"
	"time"
)

func Histogram(b *Block) {

	type histogramRule struct {
		Window int
        Key string 
	}

	type histogramData struct {
		Histogram map[string]int
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
		case query := <-b.Routes["count"]:
            data.Histogram = make(map[string]int)
            for k,pq := range(histogram){
                data.Histogram[k] = len(*pq) 
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
}
		case <-waitTimer.C:
		}
		for {
            for _
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
