package streamtools

import (
	"container/heap"
	"github.com/bitly/go-simplejson"
	"log"
	"strings"
	"time"
)

func Bunch(inChan chan *simplejson.Json, outChan chan *simplejson.Json, RuleChan chan *simplejson.Json) {

	rules := <-RuleChan
	branchString, err := rules.Get("branch").String()
	if err != nil {
		log.Fatal(err.Error())
	}
	afterSeconds, err := rules.Get("after").Int()
	if err != nil {
		log.Fatal(err.Error())
	}

	after := time.Duration(afterSeconds) * time.Second
	branch := strings.Split(branchString, ".")

	bunches := make(map[string][]*simplejson.Json)
	waitTimer := time.NewTimer(100 * time.Millisecond)
	pq := &PriorityQueue{}
	heap.Init(pq)

	for {
		select {
		case msg := <-inChan:
			id, err := msg.GetPath(branch...).String()
			if err != nil {
				log.Fatal(err.Error())
			}
			if len(bunches[id]) > 0 {
				bunches[id] = append(bunches[id], msg)
			} else {
				bunches[id] = []*simplejson.Json{msg}
			}

			val, err := simplejson.NewJson([]byte("{}"))
			if err != nil {
				log.Fatal(err.Error())
			}
			val.Set("id", id)
			val.Set("length", len(bunches[id]))

			blob, err := val.Encode()
			if err != nil {
				log.Fatal(err.Error())
			}

			queueMessage := &PQMessage{
				val: blob,
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
			queueMessage, err := simplejson.NewJson(v)
			if err != nil {
				log.Fatal(err.Error())
			}
			l, err := queueMessage.Get("length").Int()
			if err != nil {
				log.Fatal(err.Error())
			}
			id, err := queueMessage.Get("id").String()
			if err != nil {
				log.Fatal(err.Error())
			}
			if l == len(bunches[id]) {
				// we've not seen anything since putting this message in the queue
				outMsg, err := simplejson.NewJson([]byte("{}"))
				if err != nil {
					log.Fatal(err.Error())
				}
				outMsg.Set("bundle", bunches[id])
				outChan <- outMsg
			}
		}
	}
}
