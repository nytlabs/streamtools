package main

import (
    "container/heap"
    "flag"
    "github.com/bitly/go-simplejson"
    //"github.com/bitly/nsq/nsq"
    "github.com/bitly/nsq/tree/v0.2.21/nsq"
    "log"
    "strconv"
    "time"
    "net/http"
    "bytes"
    "io/ioutil"
)

var (
    // for input
    topic            = flag.String("topic", "", "nsq topic")
    channel          = flag.String("channel", "", "nsq topic")
    maxInFlight      = flag.Int("max-in-flight", 10, "max number of messages to allow in flight")
    lookupdHTTPAddrs = flag.String("lookupd-http-address", "127.0.0.1:4161", "lookupd HTTP address")
    // for output
    outNsqTCPAddrs   = flag.String("out-nsqd-tcp-address", "127.0.0.1:4151", "out nsqd TCP address")
    outTopic         = flag.String("out-topic", "", "nsq topic")
    outChannel       = flag.String("out-channel", "", "nsq channel")

    lag_time         = flag.Int("lag", 10, "lag before emitting in seconds")
    timeKey          = flag.String("key","","key that holds time")
)

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

func store(writeChan chan WriteMessage, out chan []byte, pq *PriorityQueue, lag time.Duration) {

    var emit_time time.Time
    nextMsg := &PQMessage{
        t: time.Now(),
    }

    getNext := make(chan bool)

    emitter := time.AfterFunc(24 * 365 * time.Hour, func(){
        log.Println("...")
    })

    const layout = "2006-01-02 15:04:05 -0700"

    count := 0
    heapCount := 0
    errorCount := 0 
    for {
        select {
            case inMsg := <-writeChan:
                
                outMsg := &PQMessage{
                    val:          inMsg.val,
                    t:            inMsg.t,
                }
            
                outTime := outMsg.t.Add(lag)
                outDur := outTime.Sub(time.Now()) 

                if outDur > time.Duration(0 * time.Second) {
                    heap.Push(pq, outMsg) 
                    
                    heapCount ++

                    if heapCount % 500 == 0{
                        log.Println( "HEAP: " + strconv.Itoa(pq.Len()))
                    }

                    if outMsg.t.Before(nextMsg.t) {
                        heap.Push(pq, nextMsg)
                        nextMsg = heap.Pop(pq).(*PQMessage)
                        emit_time = nextMsg.t.Add(lag)
                        duration := emit_time.Sub(time.Now()) 

                        emitter.Stop()

                        emitter = time.AfterFunc(duration, func() {
                            out<-nextMsg.val
                            count = count + 1
                            if count % 250 == 0 {
                                diff := nextMsg.t.Sub( time.Now() )
                                log.Println("POP: " + diff.String() + "IN QUEUE:" + strconv.Itoa(pq.Len()) )
                            }
                            getNext<- true
                        })
                    } 
                } else {
                    errorCount++
                    if errorCount % 250 == 0 {
                        log.Println("error: " + outDur.String() + " message reads: " + outMsg.t.Format(layout) )
                    }

                }

                //inMsg.responseChan <- true

            case <-getNext:
                if pq.Len() > 0 {
                    nextMsg = heap.Pop(pq).(*PQMessage) 
                    emit_time = nextMsg.t.Add(lag)
                    duration := emit_time.Sub(time.Now()) 

                    emitter = time.AfterFunc(duration, func() {
                        out<-nextMsg.val
                        count = count + 1
                        if count % 250 == 0 {
                            diff := nextMsg.t.Sub( time.Now() )
                            log.Println("POP: " + diff.String() + "IN QUEUE:" + strconv.Itoa(pq.Len()) )
                        }
                        getNext<- true
                    })
                }
        }
    }
}

func emitter(tcpAddr string, topic string, out chan []byte){
    outCount := 0 

    client := &http.Client{}

    for{
        select{
        case msg := <- out:
            outCount ++
            if outCount % 250 == 0{
                log.Println("OUT: " + strconv.Itoa(outCount) )
            }
            test := bytes.NewReader(msg)
            resp, err := client.Post("http://" + tcpAddr + "/put?topic=" + topic,"data/multi-part", test)
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

type SyncHandler struct{
    writeChan chan WriteMessage
    timeKey string
}

func (self *SyncHandler) HandleMessage(m *nsq.Message) error {

    reject := false

    blob, err := simplejson.NewJson(m.Body)

    if err != nil {
        reject = true
        log.Println(err.Error())
    }

    msg_time, err := blob.Get(self.timeKey).Int64()

    if err != nil {
        reject = true
        log.Println(err.Error())
    }

    // milliseconds
    t := time.Unix(0, msg_time * 1000 * 1000)
    mblob, err := blob.MarshalJSON()

    if err != nil {
        reject = true
        log.Println(err.Error())
    }

    responseChan := make(chan bool)

    msg := WriteMessage{
        t:            t,
        val:          mblob,
        responseChan: responseChan,
    }

    if !reject {
        self.writeChan <- msg
    }

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
        r, _ := nsq.NewReader(*topic, *channel)
        r.SetMaxInFlight(*maxInFlight)

        for i := 0; i < 5; i++ {
            sh := SyncHandler{
                writeChan: wc,
                timeKey: *timeKey,
            }
            r.AddHandler(&sh)
        }

        _ = r.ConnectToLookupd(*lookupdHTTPAddrs)
    }

    <-stop

}