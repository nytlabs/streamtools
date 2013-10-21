package blocks

import (
	"container/heap"
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
		case <-emitTick.C:
		case msg := <-b.InChan:
			// we should do something about the special case of "path" in the future
			// so that we only split it once, not for every message.
			keys := strings.Split(rule.Path, ".")
			msgTime := msg.GetPath(keys...).Interface().(int64)

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
				// then the queue is empty. Pause for 5 seconds before checking again
				diff = time.Duration(500) * time.Millisecond

				emitTick.Reset(diff)
				break
			}
			broadcast(b.OutChans, &item.(*PQMessage).data)
		}
	}

}
