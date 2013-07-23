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

    lateMsgCount    int
    lastStoreDiff   time.Duration
    jsonErr         int
    emitError       int
    emitCount       int
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
                    if outMsg.t.Before(nextMsg.t) {
                        heap.Push(pq, nextMsg)
                        nextMsg = heap.Pop(pq).(*PQMessage)
                        emit_time = nextMsg.t.Add(lag)
                        duration := emit_time.Sub(time.Now()) 

                        emitter.Stop()
                        emitter = time.AfterFunc(duration, func() {
                            lastStoreDiff = nextMsg.t.Sub( time.Now() )
                            out<-nextMsg.val
                            getNext<- true
                        })
                    } 
                } else {
                    lateMsgCount ++
                }

            case <-getNext:
                if pq.Len() > 0 {
                    nextMsg = heap.Pop(pq).(*PQMessage) 
                    emit_time = nextMsg.t.Add(lag)
                    duration := emit_time.Sub(time.Now()) 
                    emitter = time.AfterFunc(duration, func() {
                        lastStoreDiff = nextMsg.t.Sub( time.Now() )
                        out<-nextMsg.val
                        getNext<- true
                    })
                }
        }
    }
}

func emitter(tcpAddr string, topic string, out chan []byte){

    client := &http.Client{}

    for{
        select{
        case msg := <- out:

            msgReader := bytes.NewReader(msg)
            resp, err := client.Post("http://" + tcpAddr + "/put?topic=" + topic,"data/multi-part", msgReader)

            if err != nil {
                log.Println(err.Error())
            }

            body, err := ioutil.ReadAll(resp.Body)
            
            if string(body) != "OK" {
                log.Println(string(body))
                log.Println(err.Error())
                emitError ++ 
            } else {
                emitCount ++
            }

            resp.Body.Close()
        }
    }
}

type SyncHandler struct{
    msgChan chan *nsq.Message
    timeKey string
}

func (self *SyncHandler) HandleMessage(m *nsq.Message) error {
    self.msgChan <- m
    return nil
}

func HandleJSON(msgChan chan *nsq.Message, storeChan chan WriteMessage, timeKey string){
    for{
        select{
        case m := <- msgChan:
            blob, err := simplejson.NewJson(m.Body)
            if err != nil{
                jsonErr ++ 
                log.Println(err.Error())
                break
            }

            msgTime, err := blob.Get(timeKey).Int64()
            if err != nil{
                jsonErr ++ 
                log.Println(err.Error())
                break
            }

            ms := time.Unix(0, msgTime * 1000 * 1000)

            msg := WriteMessage{
                t:      ms,
                val:    m.Body,
            }

            storeChan <- msg
        }
    }
}


func main() {

    flag.Parse()

    wc := make(chan *nsq.Message, 500000) // SyncHandler to HandleJSON
    sc := make(chan WriteMessage)         // HandleJSON to Store
    oc := make(chan []byte)               // Store to Emitter

    lag := time.Duration(time.Duration(*lag_time) * time.Second)
    pq := &PriorityQueue{}
    heap.Init(pq)

    go HandleJSON(wc, sc, *timeKey)
    go store(sc, oc, pq, lag)
    go emitter(*outNsqTCPAddrs, *outTopic, oc)

    r, _ := nsq.NewReader(*topic, *channel)
    r.SetMaxInFlight(*maxInFlight)

    for i := 0; i < 500; i++ {
        sh := SyncHandler{
            msgChan: wc,
            timeKey: *timeKey,
        }
        r.AddHandler(&sh)
    }

    _ = r.ConnectToLookupd(*lookupdHTTPAddrs)

    go func(){
        for{
            log.Printf("\033[2J\033[1;1H")
            log.Println("messages recieved:  " + strconv.FormatUint(r.MessagesReceived, 10 ))
            log.Println("messaged finished:  " + strconv.FormatUint(r.MessagesFinished, 10 ))
            log.Println("buffered JSON chan: " + strconv.Itoa(len(wc)))
            log.Println("JSON error:         " + strconv.Itoa(jsonErr))
            log.Println("priority queue len: " + strconv.Itoa(pq.Len()))
            log.Println("store precision:    " + lastStoreDiff.String())
            log.Println("late messages:      " + strconv.Itoa(lateMsgCount))
            log.Println("emit errors:        " + strconv.Itoa(emitError))
            log.Println("emit count:         " + strconv.Itoa(emitCount))

            time.Sleep(500 * time.Millisecond)
        }
    }()

    <-r.ExitChan

}