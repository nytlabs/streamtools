package main

import (
	//"bytes"
	"container/heap"
	"flag"
	"github.com/bitly/go-simplejson"
	"github.com/bitly/nsq/tree/v0.2.21/nsq"
	//"io/ioutil"
	"log"
	//"net/http"
	"time"
	"strconv"
	"fmt"
	"runtime"
	//"bufio"
	//"net"
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

	// for logging
	lateMsgCount    int
	lastStoreDiff   time.Duration
	jsonErr         int
	emitError       int
	emitCount       int
	nextPopTime     time.Duration
	mostOff         time.Duration
	coffg5s         int
	coff1to5s       int
	coff100msto1s   int
	coff10msto100ms int
	coff1msto10ms   int
	coffl1ms        int
	s5              = time.Duration(5 * time.Second)
	s1              = time.Duration(1 * time.Second)
	ms100           = time.Duration(100 * time.Millisecond)
	ms10            = time.Duration(10 * time.Millisecond)
	ms1             = time.Duration(1 * time.Millisecond)
	pqLen           int
	lastChan    	= make(chan string)
	last  			string
)

// tallies counts of how far imprecise emitted messages are from
// their timestamp.
func OffStat(last time.Duration, lag time.Duration) {
	lastStoreDiff = last

	if lastStoreDiff < mostOff {
		mostOff = lastStoreDiff
	}

	posOff := -(lastStoreDiff + lag)

	if posOff >= s5 {
		coffg5s++
	} else if posOff < s5 && posOff >= s1 {
		coff1to5s++
	} else if posOff < s1 && posOff >= ms100 {
		coff100msto1s++
	} else if posOff < ms100 && posOff >= ms10 {
		coff10msto100ms++
	} else if posOff < ms10 && posOff >= ms1 {
		coff1msto10ms++
	} else if posOff < ms1 {
		coffl1ms++
	}
}

// checks to see if any messages on the PQ should be emitted then sends them to the emitter
func Store(in chan *PQMessage, out chan [][]byte, lag time.Duration) {
	pq := &PriorityQueue{}
	heap.Init(pq)

	emitTick := time.NewTimer(100 * time.Millisecond)
	for {
		select {
		case <-emitTick.C:
		case msg := <-in:
			lastChan <- "Store: inserting a msg"
			heap.Push(pq, msg)
			pqLen = pq.Len()
		}
		lastChan <- "Store: running peek and shift"
		now := time.Now()

		batch := make([][]byte, 0)
		msgs := 0
		for {
			item, diff := pq.PeekAndShift(now, lag)
			if item == nil {
				if !emitTick.Reset(diff) {
					emitTick = time.NewTimer(diff)
				}
				nextPopTime = diff
				break
			}
			OffStat(item.(*PQMessage).t.Sub(now), lag)
			batch = append( batch, item.(*PQMessage).val)
			msgs ++ 
			pqLen = pq.Len()
		}

		if msgs > 0{
			out <- batch
		}

		lastChan <- "Store: finished peek and shift"
	}
}

func BlockLog(){
	for{
		select {
			case m := <- lastChan:
				last = m
		}
	}
}

// accept msg from Popper and POST to NSQ
func Emitter(tcpAddr string, topic string, out chan [][]byte) {
	w := nsq.NewWriter(0)
	err := w.ConnectToNSQ(tcpAddr)

	if err != nil{
		panic(err.Error())
	}
    for {
    	select {
    	case batch := <- out:
			_, _, err := w.MultiPublish(topic, batch)
			if err != nil {
				panic(err.Error())
			}
    	}
    }

}

// synchronous handler for NSQ reader
type SyncHandler struct {
	msgChan chan *nsq.Message
	timeKey string
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
			lastChan <- "Pusher: recieved msg"

			blob, err := simplejson.NewJson(m.Body)
			if err != nil {
				jsonErr ++
				log.Println(err.Error())
				break
			}

			msgTime, err := blob.Get(timeKey).Int64()
			if err != nil {
				jsonErr ++
				log.Println(err.Error())
				break
			}

			ms := time.Unix(0, int64(time.Duration(msgTime)*time.Millisecond))
			// if this message should have already been emitted, break
			if ms.After(time.Now()) {
				lateMsgCount ++
				break
			}

			msg := &PQMessage{
				t:   ms,
				val: m.Body,
			}

			lastChan <- "Pusher: sending to store"
			store <- msg
		}
	}
}

func main() {

	flag.Parse()

	wc := make(chan *nsq.Message, 1) // SyncHandler to Pusher
	sc := make(chan *PQMessage, 1)   // Pusher to Store
	oc := make(chan [][]byte)          // Store to Emitter

	lag := time.Duration(time.Duration(*lag_time) * time.Second)
	
	numCPU := runtime.NumCPU()
	runtime.GOMAXPROCS(numCPU)

	go BlockLog()
	go Pusher(sc, wc, *timeKey, lag)           // accepts msgs from nsq handler, pushes to PQ
	go Store(sc, oc, lag)                      // pops msgs from PQ
	
	go Emitter(*outNsqTCPAddrs, *outTopic, oc) // accepts msgs from Popper, POSTs to NSQ

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

	//logging
	go func() {
		for {
			total := coffg5s +
				coff1to5s +
				coff100msto1s +
				coff10msto100ms +
				coff1msto10ms +
				coffl1ms

			offg5s := float64(coffg5s) / float64(total)
			off1to5s := float64(coff1to5s) / float64(total)
			off100msto1s := float64(coff100msto1s) / float64(total)
			off10msto100ms := float64(coff10msto100ms) / float64(total)
			off1msto10ms := float64(coff1msto10ms) / float64(total)
			offl1ms := float64(coffl1ms) / float64(total)

			logStr := fmt.Sprintf("\n"+
				"messages recieved:  %s\n"+
				"messaged finished:  %s\n"+
				"buffered JSON chan: %s\n"+
				"JSON error:         %s\n"+
				"priority queue len: %s\n"+
				"store precision:    %s\n"+
				"most imprecise:     %s\n"+
				"late messages:      %s\n"+
				"emit errors:        %s\n"+
				"emit count:         %s\n"+
				"next emit time:     %s\n"+
				"t > 5s:             %s\n"+
				"5s > t >= 1s:       %s\n"+
				"1s > t >= 100ms:    %s\n"+
				"100ms > t >= 10ms:  %s\n"+
				"10ms > t >= 1ms:    %s\n"+
				"t < 1ms:            %s\n"+
				"last message:       %s\n"+
				"NumCPU:             %s\n",
				strconv.FormatUint(r.MessagesReceived, 10),
				strconv.FormatUint(r.MessagesFinished, 10),
				strconv.Itoa(len(wc)),
				strconv.Itoa(jsonErr),
				strconv.Itoa(pqLen),
				(lag + lastStoreDiff).String(),
				(lag + mostOff).String(),
				strconv.Itoa(lateMsgCount),
				strconv.Itoa(emitError),
				strconv.Itoa(emitCount),
				nextPopTime.String(),
				strconv.FormatFloat(offg5s, 'g', 2, 64),
				strconv.FormatFloat(off1to5s, 'g', 2, 64),
				strconv.FormatFloat(off100msto1s, 'g', 2, 64),
				strconv.FormatFloat(off10msto100ms, 'g', 2, 64),
				strconv.FormatFloat(off1msto10ms, 'g', 2, 64),
				strconv.FormatFloat(offl1ms, 'g', 2, 64),
				last,
				strconv.Itoa(numCPU) )

			log.Printf("\033[2J\033[1;1H")
			log.Println(logStr)

			time.Sleep(100 * time.Millisecond)
		}
	}()

	<-r.ExitChan

}
