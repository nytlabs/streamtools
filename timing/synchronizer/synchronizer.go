package main

import (
    "container/heap"
    "flag"
    "github.com/bitly/go-simplejson"
    "github.com/bitly/nsq/tree/v0.2.21/nsq"
    "log"
    "time"
    "runtime"
    "strconv"
)

var (
    // for input
    topic            = flag.String("topic", "", "nsq topic")
    channel          = flag.String("channel", "", "nsq topic")
    maxInFlight      = flag.Int("max-in-flight", 10, "max number of messages to allow in flight")
    lookupdHTTPAddrs = flag.String("lookupd-http-address", "127.0.0.1:4161", "lookupd HTTP address")

    // for output
    outNsqTCPAddrs = flag.String("out-nsqd-tcp-address", "127.0.0.1:4150", "out nsqd TCP address")
    outTopic       = flag.String("out-topic", "", "nsq topic")
    outChannel     = flag.String("out-channel", "", "nsq channel")

    lag_time = flag.Int("lag", 10, "lag before emitting in seconds")
    timeKey  = flag.String("key", "", "key that holds time")
)

// checks to see if any messages on the PQ should be emitted then sends them to the emitter
func Store(in chan *PQMessage, out chan []byte, lag time.Duration) {
    pq := &PriorityQueue{}
    heap.Init(pq)

    emitTick := time.NewTimer(100 * time.Millisecond)
    for {
        select {
        case <-emitTick.C:
        case msg := <-in:
            heap.Push(pq, msg)
        }
        now := time.Now()
        for {
            item, diff := pq.PeekAndShift(now, lag)
            if item == nil {
                emitTick.Reset(diff)
                break
            }
            out <- item.(*PQMessage).val
        }
    }
}
// accept msg from Popper and POST to NSQ
func Emitter(tcpAddr string, topic string, out chan []byte) {
    w := nsq.NewWriter(0)
    err := w.ConnectToNSQ(tcpAddr)

    if err != nil{
        log.Fatalf(err.Error())
    }

    for {
        select {
        case msg := <- out:
            _, data, err := w.Publish(topic, msg)
            if err != nil {
                log.Println(string(data))
                log.Fatalf(err.Error())
            }
        }
    }
}

// synchronous handler for NSQ reader
type SyncHandler struct {
    msgChan chan *nsq.Message
}

func (self *SyncHandler) HandleMessage(m *nsq.Message) error {
    self.msgChan <- m
    return nil
}

// takes msg from nsq reader, parses JSON, creates a PQMessage to put in the priority queue
func Pusher(store chan *PQMessage, msgChan chan *nsq.Message, timeKey string, lag time.Duration) {
    for {
        select {
        case m := <-msgChan:

            blob, err := simplejson.NewJson(m.Body)
            if err != nil {
                log.Println(err.Error())
                break
            }

            msgTime, err := blob.Get(timeKey).Int64()
            if err != nil {
                log.Println(err.Error())
                break
            }

            ms := time.Unix(0, int64(time.Duration(msgTime)*time.Millisecond))
            // if this message should have already been emitted, break
            if ms.After(time.Now()) {
                break
            }

            msg := &PQMessage{
                t:   ms,
                val: m.Body,
            }

            store <- msg
        }
    }
}

func main() {

    flag.Parse()

    wc := make(chan *nsq.Message, 1) // SyncHandler to Pusher
    sc := make(chan *PQMessage, 1)   // Pusher to Store
    oc := make(chan []byte, 1)       // Store to Emitter

    lag := time.Duration(time.Duration(*lag_time) * time.Second)
    
    runtime.GOMAXPROCS(runtime.NumCPU())

    go Pusher(sc, wc, *timeKey, lag)           // accepts msgs from nsq handler, pushes to PQ
    go Store(sc, oc, lag)                      // pops msgs from PQ
    go Emitter(*outNsqTCPAddrs, *outTopic, oc) // accepts msgs from Popper, POSTs to NSQ

    r, _ := nsq.NewReader(*topic, *channel)
    r.SetMaxInFlight(*maxInFlight)

    sh := SyncHandler{
        msgChan: wc,
    }
    r.AddHandler(&sh)

    _ = r.ConnectToLookupd(*lookupdHTTPAddrs)

    <-r.ExitChan

}
