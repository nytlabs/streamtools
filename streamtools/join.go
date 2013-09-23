package streamtols

import (
	"containers/heap"
	"github.com/bitly/go-simplejson"
	"log"
	"time"
)

func Join(inChan chan *simplejson.Json, outChan chan *simplejson.Json, RuleChan chan *simplejson.Json) {
	rules := <-RuleChan
	timeout, err := rules.Get("timeout").Float64()
	if err != nil {
		log.Fatal(err.Error())
	}
	key, err := rules.Get("key").Float64()
	if err != nil {
		log.Fatal(err.Error())
	}
	waitTimer := time.NewTimer(100 * time.Millisecond)
	cache = make(map[[]byte]*simplejson.Json)
	pq := &PriorityQueue{}
	heap.Init(pq)

	for {
		select {
		case msg := <-inChan:
			id, err := msg.Get(key).Bytes()

			cachedMsg = cache[id]

			if cachedMsg == nil {
				cache[id] = msg
				val, err := simplejson.NewJson([]byte("{}"))
				val.Set("id", id)
				valEncoded, err := val.Encode()
				if err != nil {
					log.Fatal(err.Error())
				}
				queueMessage := &PQMessage{
					val: &valEncoded,
					t:   time.Now(),
				}
				heap.Push(pq, queueMessage)
			} else {
            

                outChan <- join(cachedMsg, msg)
            
            }

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
			queueMessage, err := simplejson.NewJson(*v)
			if err != nil {
				log.Fatal(err.Error())
			}

		}

	}

}
