package blocks

import (
	"container/heap"
	"github.com/nytlabs/gojee"
	"log"
	"time"
)

func pqAverage(pq *PriorityQueue) float64 {
	var sum float64
	sum = 0
	for _, pqmsg := range *pq {
		v := pqmsg.val
		val, ok := v.(float64)
		if !ok {
			log.Println("non float stored in moving average")
			continue
		}
		sum += val
	}
	N := float64(len(*pq))
	return sum / N
}

func MovingAverage(b *Block) {

	type movingAverageRule struct {
		Window string
		Path   string
	}

	type avgData struct {
		Average float64
	}

	var rule *movingAverageRule
	var tree *jee.TokenTree

	window := time.Duration(0)
	waitTimer := time.NewTimer(100 * time.Millisecond)

	pq := &PriorityQueue{}
	heap.Init(pq)

	for {
		select {
		case query := <-b.Routes["moving_average"]:
			data := avgData{
				Average: pqAverage(pq),
			}
			marshal(query, data)
		case <-b.Routes["poll"]:
			outMsg := map[string]interface{}{
				"Averageg": pqAverage(pq),
			}
			out := BMsg{
				Msg: outMsg,
			}
			broadcast(b.OutChans, &out)
		case ruleUpdate := <-b.Routes["set_rule"]:
			if rule == nil {
				rule = &movingAverageRule{}
			}
			unmarshal(ruleUpdate, rule)
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
			window, err = time.ParseDuration(rule.Window)
			if err != nil {
				log.Println(err.Error())
			}
		case msg := <-b.Routes["get_rule"]:
			if rule == nil {
				marshal(msg, &movingAverageRule{})
			} else {
				marshal(msg, rule)
			}
		case <-b.QuitChan:
			quit(b)
			return
		case msg := <-b.AddChan:
			updateOutChans(msg, b)
		case msg := <-b.InChan:
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
			// TODO make this a type swtich and convert anything we can to a
			// float
			val, ok := val.(float64)
			if !ok {
				log.Println("trying to put a non-float into the moving average")
				continue
			}
			queueMessage := &PQMessage{
				val: val,
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
