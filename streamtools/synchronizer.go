package streamtools

import (
	"container/heap"
	"github.com/bitly/go-simplejson"
	"log"
	"time"
)

func Synchronizer(inChan chan *simplejson.Json, outChan chan *simplejson.Json, RuleChan chan *simplejson.Json) {
	pq := &PriorityQueue{}
	heap.Init(pq)

	params := <-RuleChan

	timeKey, err := params.Get("key").String()
	if err != nil {
		log.Println(err.Error())
	}

	lag, err := params.Get("lag").Int()
	if err != nil {
		log.Println(err.Error())
	}

	lagTime := time.Duration(time.Duration(lag) * time.Second)

	emitTick := time.NewTimer(100 * time.Millisecond)

	for {
		select {
		case <-RuleChan:
		case <-emitTick.C:
		case msg := <-inChan:
			msgTime, err := msg.Get(timeKey).Int64()
			if err != nil {
				log.Fatalf(err.Error())
			}

			// assuming the value is in MS
			// TODO: make this more explicit and/or flexible
			ms := time.Unix(0, int64(time.Duration(msgTime)*time.Millisecond))

			queueMessage := &PQMessage{
				data: *msg,
				t:    ms,
			}

			heap.Push(pq, queueMessage)
		}
		now := time.Now()
		for {
			item, diff := pq.PeekAndShift(now, lagTime)
			if item == nil {
				emitTick.Reset(diff)
				break
			}

			outChan <- &item.(*PQMessage).data
		}
	}

}
