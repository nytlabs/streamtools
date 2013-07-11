package main

import (
	"container/heap"
	"flag"
	"github.com/bitly/go-simplejson"
	"github.com/bitly/nsq/nsq"
	"log"
	"strconv"
	"time"
)

var (
	topic            = flag.String("topic", "", "nsq topic")
	channel          = flag.String("channel", "", "nsq topic")
	maxInFlight      = flag.Int("max-in-flight", 1000, "max number of messages to allow in flight")
	nsqTCPAddrs      = flag.String("nsqd-tcp-address", "", "nsqd TCP address")
	nsqHTTPAddrs     = flag.String("nsqd-http-address", "", "nsqd HTTP address")
	lookupdHTTPAddrs = flag.String("lookupd-http-address", "", "lookupd HTTP address")
	lag_time         = flag.Int("lag", 10, "lag before emitting in seconds")
)

// MESSAGE HANDLER FOR THE NSQ READER
type MessageHandler struct {
	msgChan  chan *nsq.Message
	stopChan chan int
}

func (self *MessageHandler) HandleMessage(message *nsq.Message, responseChannel chan *nsq.FinishedMessage) {
	self.msgChan <- message
	responseChannel <- &nsq.FinishedMessage{message.Id, 0, true}
}

type WriteMessage struct {
	val          []byte
	t            time.Time
	responseChan chan bool
}

type PQMessage struct {
	val          []byte
	t            time.Time
	index        int
	killChan     chan bool
	responseChan chan bool
}

// PRIORITY QUEUE
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

// update modifies the priority and value of an Item in the queue.
func (pq *PriorityQueue) update(item *PQMessage, val []byte, time time.Time) {
	heap.Remove(pq, item.index)
	item.val = val
	item.t = time
	heap.Push(pq, item)
}

func emit(msg *PQMessage, lag time.Duration) {
	emit_at := msg.t.Add(lag)
	log.Println("Sleeping")
	time.Sleep(emit_at.Sub(time.Now()))
	select {
	case <-msg.killChan:
		// do nowt  
		log.Printf("killed")
	default:
		log.Printf(
			"### item's timestamp: %s. emitted at %s \n",
			msg.t.Format("15:04:05"),
			time.Now().Format("15:04:05"),
		)
	}
	msg.responseChan <- true
}

func store(writeChan chan WriteMessage, pq *PriorityQueue, lag time.Duration) {

	nextMsg := &PQMessage{
		responseChan: make(chan bool),
	}
	nextMsgChan := make(chan *PQMessage, 1)
	getNextMsg := true

	for {
		// if a message is ready, pop it and make it ready for emitting
		if pq.Len() > 0 && getNextMsg {
			nextMsg = heap.Pop(pq).(*PQMessage)
			nextMsgChan <- nextMsg
			// let's not get another message until this one is done
			getNextMsg = false
		}

		select {
		case inMsg := <-writeChan:
			// if we've recieved something in the write channel, push it onto the heap
			outMsg := &PQMessage{
				val:          inMsg.val,
				t:            inMsg.t,
				killChan:     make(chan bool, 1),
				responseChan: make(chan bool),
			}
			heap.Push(pq, outMsg)

			// check it didn't arrive before the current next message
			if nextMsg.val != nil {
				if outMsg.t.Before(nextMsg.t) {
					// if it did, kill the current emitter
					nextMsg.killChan <- true
					// put the old one back in the queue
					heap.Push(pq, nextMsg)
					// make sure the new message is loaded on the next iteration
					getNextMsg = true
				}
			}
			inMsg.responseChan <- true
		case outMsg := <-nextMsgChan:
			// if something is waitingto be emitted set it going
			go emit(outMsg, lag)
		case <-nextMsg.responseChan:
			// make sure the next message is loaded when the current nextMsg is done
			getNextMsg = true

		}
	}
}

// function to read an NSQ channel and write to the key value store
func writer(mh MessageHandler, writeChan chan WriteMessage) {
	for {
		select {
		case m := <-mh.msgChan:

			blob, err := simplejson.NewJson(m.Body)

			if err != nil {
				log.Fatalf(err.Error())
			}

			val, err := blob.Get("t").String()

			if err != nil {
				log.Fatalf(err.Error())
			}

			msg_time, err := strconv.ParseInt(val, 0, 64)

			if err != nil {
				log.Fatalf(err.Error())
			}

			t := time.Unix(0, msg_time)
			mblob, err := blob.MarshalJSON()

			if err != nil {
				log.Fatalf(err.Error())
			}

			responseChan := make(chan bool)

			msg := WriteMessage{
				t:            t,
				val:          mblob,
				responseChan: responseChan,
			}

			writeChan <- msg

			success := <-responseChan

			if !success {
				// TODO learn about err.Error()
				log.Fatalf("its broken")
			}
		}
	}
}

func main() {

	flag.Parse()

	r, err := nsq.NewReader(*topic, *channel)

	if err != nil {
		log.Fatal(err.Error())
	}

	mh := MessageHandler{
		msgChan:  make(chan *nsq.Message, 5),
		stopChan: make(chan int),
	}

	wc := make(chan WriteMessage)

	pq := &PriorityQueue{}
	heap.Init(pq)

	lag := time.Duration(time.Duration(*lag_time) * time.Second)

	go store(wc, pq, lag)
	go writer(mh, wc)

	r.AddAsyncHandler(&mh)

	err = r.ConnectToNSQ(*nsqTCPAddrs)
	if err != nil {
		log.Fatalf(err.Error())
	}
	err = r.ConnectToLookupd(*lookupdHTTPAddrs)
	if err != nil {
		log.Fatalf(err.Error())
	}

	<-mh.stopChan
}
