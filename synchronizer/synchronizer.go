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
    nextPopTime     time.Time
    firingTimeDiff  time.Duration
)

type WriteMessage struct {
	val          []byte
	t            time.Time
	responseChan chan bool
}

func store(inChan chan WriteMessage, outChan chan []byte, pq *PriorityQueue, lag time.Duration) {

    sleepTimer := time.NewTimer(time.Duration(0))
    var outMsg interface{}
    outMsg = nil

    for {
        select {
        case <-sleepTimer.C:
            if outMsg != nil {
                outChan <- outMsg.(*PQMessage).val
                lastStoreDiff = outMsg.(*PQMessage).t.Sub( time.Now() )
            }

            if pq.Len() > 0 {
                outMsg = heap.Pop(pq).(*PQMessage)
                emit_time := outMsg.(*PQMessage).t.Add(lag)
                nextPopTime = emit_time
                duration := emit_time.Sub(time.Now())
                sleepTimer.Reset(duration)
            } else {
                outMsg = nil
            }

        case msg := <-inChan:
            qMsg := &PQMessage{
                val: msg.val,
                t:   msg.t,
            }

            outTime := qMsg.t.Add(lag)
            outDur := outTime.Sub(time.Now()) 

            if outDur > time.Duration(0 * time.Second) {

                if outMsg == nil || qMsg.t.Before(outMsg.(*PQMessage).t) {
                    emit_time := qMsg.t.Add(lag)
                    nextPopTime = emit_time
                    duration := emit_time.Sub(time.Now())
                    sleepTimer.Reset(duration)
                    if outMsg != nil{
                        heap.Push(pq, outMsg)
                    }
                    outMsg = qMsg
                } else {
                    heap.Push(pq, qMsg)
                }

            } else {
                lateMsgCount ++ 
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
                log.Fatalf(err.Error())
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
    const layout = "Jan 2, 2006 at 3:04pm (MST)"

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

    for i := 0; i < 5; i++ {
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
            log.Println("firing time diff:   " + firingTimeDiff.String() )
            log.Println("late messages:      " + strconv.Itoa(lateMsgCount))
            log.Println("emit errors:        " + strconv.Itoa(emitError))
            log.Println("emit count:         " + strconv.Itoa(emitCount))
            log.Println("next emit time:     " + nextPopTime.Format(layout))
            time.Sleep(100 * time.Millisecond)
        }
    }()

    <-r.ExitChan

}