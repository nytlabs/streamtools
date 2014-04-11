package library

// PRIORITY QUEUE

import (
	"container/heap"
	"time"
)

type PQMessage struct {
	val   interface{}
	t     time.Time
	index int
}

// A PriorityQueue implements heap.Interface and holds Items.
type PriorityQueue []*PQMessage

func (pq PriorityQueue) Len() int {
	return len(pq)
}

func (pq PriorityQueue) Less(i, j int) bool {
	// We want Pop to give us the highest, not lowest, priority so we use greater than here.
	return pq[i].t.Before(pq[j].t)
}

func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func (pq *PriorityQueue) Push(x interface{}) {
	n := len(*pq)
	item := x.(*PQMessage)
	item.index = n
	*pq = append(*pq, item)
}

func (pq *PriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	item.index = -1 // for safety
	*pq = old[0 : n-1]
	return item
}

func (pq *PriorityQueue) Peek() interface{} {
	if pq.Len() == 0 {
		return nil
	}
	return (*pq)[0]
}

func (pq *PriorityQueue) PeekAndShift(max time.Time, lag time.Duration) (interface{}, time.Duration) {
	if pq.Len() == 0 {
		return nil, 0
	}

	item := (*pq)[0]

	if item.t.Add(lag).Before(max) {
		heap.Remove(pq, 0)
		return item, 0
	}

	return nil, lag - max.Sub(item.t)
}
