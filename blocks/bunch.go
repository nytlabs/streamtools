package blocks

import (
	"container/heap"
	"log"
	"strings"
	"time"
)

func Bunch(b *Block) {

	type bunchRule struct {
		Branch    string
		EmitAfter int
	}

	var rule *bunchRule
	var branch []string

	after := time.Duration(0)
	waitTimer := time.NewTimer(100 * time.Millisecond)
	bunches := make(map[string][]BMsg)

	pq := &PriorityQueue{}
	heap.Init(pq)

	for {
		select {
		case msg := <-b.AddChan:
			updateOutChans(msg, b)
		case <-b.QuitChan:
			quit(b)
			return
		case msg := <-b.Routes["set_rule"]:
			if rule == nil {
				rule = &bunchRule{}
			}
			unmarshal(msg, rule)
			after = time.Duration(rule.EmitAfter) * time.Second
			branch = strings.Split(rule.Branch, ".")
		case msg := <-b.Routes["get_rule"]:
			if rule == nil {
				marshal(msg, &bunchRule{})
			} else {
				marshal(msg, rule)
			}
		case msg := <-b.InChan:
			if rule == nil {
				break
			}

			id, err := Get(msg, branch...)
			idStr, ok := id.(string)
			if !ok {
				log.Fatal("type assertion failed")
			}
			if err != nil {
				log.Fatal(err.Error())
			}
			if len(bunches[idStr]) > 0 {
				bunches[idStr] = append(bunches[idStr], msg)
			} else {
				bunches[idStr] = []BMsg{msg}
			}

			val := make(map[string]interface{})
			err = Set(val, "id", idStr)
			if err != nil {
				log.Fatal(err.Error())
			}
			err = Set(val, "length", len(bunches[idStr]))
			if err != nil {
				log.Fatal(err.Error())
			}
			queueMessage := &PQMessage{
				val: val,
				t:   time.Now(),
			}
			heap.Push(pq, queueMessage)
		case <-waitTimer.C:
		}
		for {
			pqMsg, diff := pq.PeekAndShift(time.Now(), after)
			if pqMsg == nil {
				// either the queue is empty, or it's not time to emit
				waitTimer.Reset(diff)
				break
			}
			v := pqMsg.(*PQMessage).val

			l, err := Get(v, "length")
			if err != nil {
				log.Fatal(err.Error())
			}
			lInt := l.(int)
			id, err := Get(v, "id")
			if err != nil {
				log.Fatal(err.Error())
			}
			idStr := id.(string)
			if lInt == len(bunches[idStr]) {
				// we've not seen anything since putting this message in the queue
				outMsg := make(map[string]interface{})
				err = Set(outMsg, "bunch", bunches[idStr])
				if err != nil {
					log.Fatal(err.Error())
				}
				broadcast(b.OutChans, outMsg)
				delete(bunches, idStr)
			}
		}
	}
}
