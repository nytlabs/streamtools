package blocks

import (
	"container/heap"
	"github.com/nytlabs/gojee"
	"log"
	"time"
)

// Count uses a priority queue to count the number of messages that have been sent
// to the count block over a duration of time in seconds.
//
// Note that this is an exact count and therefore has O(N) memory requirements.
func MovingAverage(b *Block) {

	var err error

	type movingAverageRule struct {
		Window string
		Path   string
	}

	type avgData struct {
		Average int
		Window  string
	}

	data := &avgData{Count: 0}
	var rule *countRule
	var tree *jee.TokenTree

	window := time.Duration(0)
	waitTimer := time.NewTimer(100 * time.Millisecond)

	pq := &PriorityQueue{}
	heap.Init(pq)

	emptyByte := make([]byte, 0)

	for {
		select {
		case query := <-b.Routes["moving_average"]:
			data.Count = len(*pq)
			marshal(query, data)
		case <-b.Routes["poll"]:
			outMsg := map[string]interface{}{
				"Count": len(*pq),
			}
			out := BMsg{
				Msg: outMsg,
			}
			broadcast(b.OutChans, &out)
		case ruleUpdate := <-b.Routes["set_rule"]:
			if rule == nil {
				rule = &countRule{}
			}
			unmarshal(ruleUpdate, rule)
			token, err := jee.Lexer(rule.Key)
			if err != nil {
				log.Println(err.Error())
				break
			}
			tree, err = jee.Parser(token)
			if err != nil {
				log.Println(err.Error())
				break
			}
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
			if tree == nil {
				break
			}
			val, err := jee.Eval(tree, msg.Msg)
			if err != nil {
				log.Println(err.Error())
				break
			}

			queueMessage := &PQMessage{
				val: &val,
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
