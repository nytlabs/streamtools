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
    "fmt"
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
    nextPopTime     time.Duration
    mostOff         time.Duration

    coffg5s          int
    coff1to5s        int
    coff100msto1s    int
    coff10msto100ms  int
    coff1msto10ms    int
    coffl1ms         int
    s5              = time.Duration(5 * time.Second)
    s1              = time.Duration(1 * time.Second)
    ms100           = time.Duration(100 * time.Millisecond)
    ms10            = time.Duration(10 * time.Millisecond)
    ms1             = time.Duration(1 * time.Millisecond)
    pqLen            int
)

// tallies counts of how far imprecise emitted messages are from 
// their timestamp. 
func OffStat(last time.Duration, lag time.Duration){
    lastStoreDiff = last

    if lastStoreDiff < mostOff {
        mostOff = lastStoreDiff
    }

    posOff := -(lastStoreDiff + lag) 

    if posOff >= s5 {
        coffg5s ++ 
    } else if posOff < s5 && posOff >= s1 {
        coff1to5s++
    } else if posOff < s1 && posOff >= ms100 {
        coff100msto1s ++
    } else if posOff < ms100 && posOff >= ms10 {
        coff10msto100ms ++
    } else if posOff < ms10 && posOff >= ms1{
        coff1msto10ms ++
    } else if posOff < ms1 {
        coffl1ms ++ 
    }
}

// checks to see if any messages on the PQ should be emitted then sends them to the emitter
func Store(in chan *PQMessage, out chan []byte, lag time.Duration){
    pq := &PriorityQueue{}
    heap.Init(pq)

    emitTick := time.NewTimer(100 * time.Hour)
    emitTime := time.Now().Add(100 * time.Hour)
    for {
        select{
            case _ = <- emitTick.C:
                outMsg := heap.Pop(pq).(*PQMessage)
                OffStat( outMsg.t.Sub( time.Now() ), lag )
                out <- outMsg.val

                if pq.Len() > 0 {
                    delay := lag - time.Now().Sub( pq.Peek().(*PQMessage).t ) 
                    emitTick.Reset( delay )
                    nextPopTime = delay
                    emitTime = time.Now().Add(delay)
                }

                pqLen = pq.Len()

            case msg := <- in:
                heap.Push(pq, msg)
                if pq.Peek().(*PQMessage).t.Before(emitTime) {
                    delay := lag - time.Now().Sub( pq.Peek().(*PQMessage).t ) 
                    emitTick.Reset( delay )
                    nextPopTime = delay
                    emitTime = time.Now().Add(delay)
                }
        }
    }

}

// accept msg from Popper and POST to NSQ
func Emitter(tcpAddr string, topic string, out chan []byte){
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
                log.Println(err.Error())
                emitError ++ 
            } else {
                emitCount ++
            }

            resp.Body.Close()
        }
    }
}

// synchronous handler for NSQ reader
type SyncHandler struct{
    msgChan chan *nsq.Message
    timeKey string
}

func (self *SyncHandler) HandleMessage(m *nsq.Message) error {
    self.msgChan <- m
    return nil
}

// takes msg from nsq reader, parses JSON, creates a PQMessage to put in the priority queue
func Pusher(store chan *PQMessage, msgChan chan *nsq.Message, timeKey string, lag time.Duration){
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

            msg := &PQMessage{
                t:      ms,
                val:    m.Body,
            }

            // block for a short time so that we don't render the PQ useless 
            // and block our msg emitting
            time.Sleep(100 * time.Microsecond)

            outTime := msg.t.Add(lag)
            outDur := outTime.Sub(time.Now()) 

            // if this message shouldn't have already been emitted, add to PQ
            if outDur > time.Duration( 0 * time.Second) {
                //heap.Push(pq, msg)
                store <- msg
            } else {
                lateMsgCount ++ 
            }
        }
    }
}

func main() {
    const layout = "Jan 2, 2006 at 3:04pm (MST)"

    flag.Parse()

    wc := make(chan *nsq.Message, 500000) // SyncHandler to Pusher
    oc := make(chan []byte)               // Store to Emitter
    sc := make(chan *PQMessage,1000)

    lag := time.Duration(time.Duration(*lag_time) * time.Second)

    go Pusher(sc, wc, *timeKey, lag)            // accepts msgs from nsq handler, pushes to PQ
    go Store(sc, oc, lag)                      // pops msgs from PQ
    go Emitter(*outNsqTCPAddrs, *outTopic, oc)  // accepts msgs from Popper, POSTs to NSQ

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
            total := coffg5s          +
                     coff1to5s        +
                     coff100msto1s    +
                     coff10msto100ms  +
                     coff1msto10ms    +
                     coffl1ms         

            offg5s          := float64(coffg5s)          / float64(total)
            off1to5s        := float64(coff1to5s)        / float64(total)
            off100msto1s    := float64(coff100msto1s)    / float64(total)
            off10msto100ms  := float64(coff10msto100ms)  / float64(total)
            off1msto10ms    := float64(coff1msto10ms)    / float64(total)
            offl1ms         := float64(coffl1ms)         / float64(total)

            logStr := fmt.Sprintf("\n"+
                "messages recieved:  %s\n" +
                "messaged finished:  %s\n" +
                "buffered JSON chan: %s\n" +
                "JSON error:         %s\n" +
                "priority queue len: %s\n" +
                "store precision:    %s\n" +
                "most imprecise:     %s\n" +
                "late messages:      %s\n" +
                "emit errors:        %s\n" +
                "emit count:         %s\n" +
                "next emit time:     %s\n" +
                "t > 5s:             %s\n" +
                "5s > t >= 1s:       %s\n" + 
                "1s > t >= 100ms:    %s\n" + 
                "100ms > t >= 10ms:  %s\n" + 
                "10ms > t >= 1ms:    %s\n" + 
                "t < 1ms:            %s\n",
                strconv.FormatUint(r.MessagesReceived, 10 ), 
                strconv.FormatUint(r.MessagesFinished, 10 ),
                strconv.Itoa(len(wc)),
                strconv.Itoa(jsonErr),
                strconv.Itoa(pqLen), 
                (lag + lastStoreDiff).String(), 
                (lag + mostOff).String(),
                strconv.Itoa(lateMsgCount),
                strconv.Itoa(emitError),
                strconv.Itoa(emitCount),
                nextPopTime.String(),
                strconv.FormatFloat(offg5s,'g',2, 64),
                strconv.FormatFloat(off1to5s,'g',2, 64),
                strconv.FormatFloat(off100msto1s,'g',2, 64),
                strconv.FormatFloat(off10msto100ms,'g',2, 64),
                strconv.FormatFloat(off1msto10ms,'g',2, 64),
                strconv.FormatFloat(offl1ms,'g',2, 64))
    
            log.Printf("\033[2J\033[1;1H")
            log.Println(logStr)

            time.Sleep(100 * time.Millisecond)
        }
    }()

    <-r.ExitChan

}