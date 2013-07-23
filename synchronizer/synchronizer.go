package main

import (
	"bytes"
	"container/heap"
	"flag"
	"github.com/bitly/go-simplejson"
	"github.com/bitly/nsq/nsq"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"
)

var (
	// for input
	topic            = flag.String("topic", "", "nsq topic")
	channel          = flag.String("channel", "", "nsq topic")
	maxInFlight      = flag.Int("max-in-flight", 10, "max number of messages to allow in flight")
	lookupdHTTPAddrs = flag.String("lookupd-http-address", "127.0.0.1:4161", "lookupd HTTP address")
	// for output
	outNsqTCPAddrs = flag.String("out-nsqd-tcp-address", "127.0.0.1:4151", "out nsqd TCP address")
	outTopic       = flag.String("out-topic", "", "nsq topic")
	outChannel     = flag.String("out-channel", "", "nsq channel")

	lag_time = flag.Int("lag", 10, "lag before emitting in seconds")
	timeKey  = flag.String("key", "", "key that holds time")
)

type WriteMessage struct {
	val          []byte
	t            time.Time
	responseChan chan bool
}

func store(inChan chan WriteMessage, outChan chan []byte, pq *PriorityQueue, lag time.Duration) {

	var sleepTimer *time.Timer

	for {
		select {
		case <-sleepTimer.C:
			outMsg := heap.Pop(pq).(*PQMessage)
			outChan <- outMsg.val

		case msg := <-inChan:
			qMsg := &PQMessage{
				val: msg.val,
				t:   msg.t,
			}
			head_msg := pq.Peek().(*PQMessage)
			if head_msg == nil {
				// if we don't have anything on the pqueue
				// then stick this one on the queue
				heap.Push(pq, qMsg)
				// and start a timer
				emit_time := msg.t.Add(lag)
				duration := emit_time.Sub(time.Now())
				sleepTimer = time.NewTimer(duration)
			} else {
				// if there are things on the pqueue
				// check to see if we need to push onto the front of the queue
				if head_msg.t.After(msg.t) {
					// reset the timer
					emit_time := msg.t.Add(lag)
					duration := emit_time.Sub(time.Now())
					was_active := sleepTimer.Reset(duration)
					if was_active {
						// if we're not too late, push the msg onto the pqueue
						heap.Push(pq, qMsg)
					} else {
						log.Fatal("we're too late!")
					}
				} else {
					heap.Push(pq, qMsg)
				}
			}
		}
	}
}

func emitter(tcpAddr string, topic string, out chan []byte) {
	outCount := 0

	client := &http.Client{}

	for {
		select {
		case msg := <-out:
			outCount++
			if outCount%250 == 0 {
				log.Println("OUT: " + strconv.Itoa(outCount))
			}
			test := bytes.NewReader(msg)
			resp, err := client.Post("http://"+tcpAddr+"/put?topic="+topic, "data/multi-part", test)
			if err != nil {
				log.Println(err.Error())
			}
			body, err := ioutil.ReadAll(resp.Body)

			if string(body) != "OK" {
				log.Println(body)
			}

			resp.Body.Close()
		}
	}
}

type SyncHandler struct {
	writeChan chan WriteMessage
	timeKey   string
}

func (self *SyncHandler) HandleMessage(m *nsq.Message) error {

	blob, err := simplejson.NewJson(m.Body)

	if err != nil {
		log.Println(err.Error())
		return nil
	}

	msg_time, err := blob.Get(self.timeKey).Int64()

	if err != nil {
		log.Println(err.Error())
		return nil
	}

	// milliseconds
	t := time.Unix(0, msg_time*1000*1000)
	mblob, err := blob.MarshalJSON()

	if err != nil {
		log.Println(err.Error())
		return nil
	}

	responseChan := make(chan bool)

	msg := WriteMessage{
		t:            t,
		val:          mblob,
		responseChan: responseChan,
	}

	self.writeChan <- msg

	return nil
}

func main() {

	flag.Parse()

	stop := make(chan bool)
	wc := make(chan WriteMessage, 500000)
	oc := make(chan []byte)
	pq := &PriorityQueue{}
	heap.Init(pq)

	lag := time.Duration(time.Duration(*lag_time) * time.Second)

	go store(wc, oc, pq, lag)
	go emitter(*outNsqTCPAddrs, *outTopic, oc)

	for j := 0; j < 10; j++ {
		go func() {
			r, _ := nsq.NewReader(*topic, *channel)
			r.SetMaxInFlight(*maxInFlight)

			for i := 0; i < 5; i++ {
				sh := SyncHandler{
					writeChan: wc,
					timeKey:   *timeKey,
				}
				r.AddHandler(&sh)
			}

			_ = r.ConnectToLookupd(*lookupdHTTPAddrs)
		}()
	}

	<-stop

}
