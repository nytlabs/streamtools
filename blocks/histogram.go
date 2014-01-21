package blocks

import (
	"container/heap"
	"github.com/nytlabs/gojee" // jee
	"log"
	"time"
)

// build the histogram JSON for output
func buildHistogram(histogram map[string]*PriorityQueue) map[string]interface{} {

	buckets := make([]map[string]interface{}, len(histogram))

	i := 0
	for k, pq := range histogram {
		bucket := map[string]interface{}{
			"Count": len(*pq),
			"Label": k,
		}
		buckets[i] = bucket
		i++
	}
	data := map[string]interface{}{
		"Histogram": buckets,
	}
	return data
}

// creates a histogram of a specified key
func Histogram(b *Block) {

	type histogramRule struct {
		Window string
		Path   string
	}

	var rule *histogramRule
	var tree *jee.TokenTree
	var err error

	waitTimer := time.NewTimer(100 * time.Millisecond)
	window := time.Duration(0)

	histogram := map[string]*PriorityQueue{}
	emptyByte := make([]byte, 0)

	for {
		select {
		case query := <-b.Routes["histogram"]:
			data := buildHistogram(histogram)
			marshal(query, data)
		case <-b.Routes["poll"]:
			data := buildHistogram(histogram)
			out := BMsg{
				Msg: data,
			}
			broadcast(b.OutChans, out)
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
			window, err = time.ParseDuration(rule.Window)
			if err != nil {
				log.Println(err.Error())
			}
			token, err := jee.Lexer(rule.Path)
			if err != nil {
				log.Println(err.Error())
				break
			}
			tree, err = jee.Parser(token)
			if err != nil {
				log.Println(err.Error())
				break
			}
		case msg := <-b.AddChan:
			updateOutChans(msg, b)
		case msg := <-b.InChan:
			if rule == nil {
				break
			}

			v, err := jee.Eval(tree, msg.Msg)
			if err != nil {
				log.Println(err.Error())
				break
			}
			valueString, ok := v.(string)
			if !ok {
				log.Println("nil value against", rule.Path, " - ignoring")
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
