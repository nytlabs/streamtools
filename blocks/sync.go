package blocks

import (
	"container/heap"
	"log"
	"strings"
	"time"
)

func Sync(b *Block) {
	pq := &PriorityQueue{}
	heap.Init(pq)

	type syncRule struct {
		Path string
		Lag  int
	}

	rule := &syncRule{}

	unmarshal(<-b.Routes["set_rule"], &rule)

	lagTime := time.Duration(time.Duration(rule.Lag) * time.Second)

	emitTick := time.NewTimer(500 * time.Millisecond)

	for {
		select {
		case m := <-b.Routes["set_rule"]:
			unmarshal(m, &rule)
			lagTime = time.Duration(time.Duration(rule.Lag) * time.Second)
		case m := <-b.Routes["get_rule"]:
			marshal(m, rule)
		case msg := <-b.AddChan:
			updateOutChans(msg, b)
		case <-b.QuitChan:
			quit(b)
			return
		case <-emitTick.C:
		case msg := <-b.InChan:
			keys := strings.Split(rule.Path, ".")
			msgTime, err := Get(msg, keys...)
			if err != nil {
				log.Println(err.Error())
			}
			msgTimeI, ok := msgTime.(int64)
			if !ok {
				log.Println("could not cast time key to int")
			}

			// assuming the value is in MS
			// TODO: make this more explicit and/or flexible
			log.Println(msgTimeI)
			ms := time.Unix(0, int64(time.Duration(msgTimeI)*time.Millisecond))
			log.Println(ms)

			queueMessage := &PQMessage{
				val:  msg,
				t:    ms,
			}

			heap.Push(pq, queueMessage)
		}
		now := time.Now()
		for {
			item, diff := pq.PeekAndShift(now, lagTime)
			if item == nil {
				// then the queue is empty. Pause for 5 seconds before checking again
				diff = time.Duration(500) * time.Millisecond

				emitTick.Reset(diff)
				break
			}
			broadcast(b.OutChans, &item.(*PQMessage).val)
		}
	}

}
