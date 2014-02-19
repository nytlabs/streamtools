package blocks

import (
	"container/heap"
	"github.com/nytlabs/gojee"
	"log"
	"time"
)

// Pack groups messages together by testing equality of keys, and emits the
// group of messages after no new messages have been added to the group for a
// specified amount of time.
func Pack(b *Block) {

	type bunchRule struct {
		Path      string
		EmitAfter string
	}

	var rule *bunchRule
	var tree *jee.TokenTree

	after := time.Duration(0)
	waitTimer := time.NewTimer(100 * time.Millisecond)
	bunches := make(map[string][]interface{})

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

			after, err = time.ParseDuration(rule.EmitAfter)
			if err != nil {
				log.Println(err.Error())
				break
			}
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
			if tree == nil {
				break
			}

			id, err := jee.Eval(tree, msg.Msg)
			if err != nil {
				log.Println(err.Error())
				break
			}
			idStr, ok := id.(string)
			if !ok {
				log.Println("could not assert id to string")
				break
			}
			if len(bunches[idStr]) > 0 {
				bunches[idStr] = append(bunches[idStr], msg.Msg)
			} else {
				bunches[idStr] = []interface{}{msg.Msg}
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
				msg := make(map[string]interface{})
				err = Set(msg, "bunch", bunches[idStr])
				if err != nil {
					log.Fatal(err.Error())
				}
				outMsg := BMsg{
					Msg:          msg,
					ResponseChan: nil,
				}

				broadcast(b.OutChans, &outMsg)
				delete(bunches, idStr)
			}
		}
	}
}
