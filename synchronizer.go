package main

import (
    "log"
    "github.com/bitly/nsq/nsq"
    "github.com/bitly/go-simplejson"
    "flag"
    "strconv"
    "time"
    "fmt"
)

var (
    topic = flag.String("topic", "", "nsq topic")
    channel = flag.String("channel", "", "nsq topic")
    maxInFlight = flag.Int("max-in-flight", 1000, "max number of messages to allow in flight")
    nsqTCPAddrs = flag.String("nsqd-tcp-address", "", "nsqd TCP address")
    nsqHTTPAddrs = flag.String("nsqd-http-address", "", "nsqd HTTP address")
    lookupdHTTPAddrs = flag.String("lookupd-http-address", "", "lookupd HTTP address")
)

type MessageHandler struct {
    msgChan chan *nsq.Message
    stopChan chan int
}

func (self *MessageHandler) HandleMessage(message *nsq.Message, responseChannel chan *nsq.FinishedMessage){
    self.msgChan <- message
    responseChannel <- &nsq.FinishedMessage{message.Id, 0, true}
}

type SyncMessage struct{
    message []byte
    t uint64
}

type WriteMessage struct{
    key uint64 // bucket time
    val []byte
    t uint64   // msg time
    responseChan chan bool
}

type ReadMessage struct{
    key uint64
    responseChan chan []SyncMessage
}

func emitter(readChan chan ReadMessage, lag_time uint64){
    c:= time.Tick(40 * time.Millisecond)
    responseChan := make(chan []SyncMessage)

    for now := range c {
        
        cur_time := uint64( now.UnixNano() / 1000000 )

        readMsg := ReadMessage {
            key: cur_time - cur_time % 40 - lag_time,
            responseChan: responseChan,
        }

        readChan <- readMsg

        msgs := <- responseChan

        if len(msgs) > 0 {
            fmt.Printf("recieved %d for %d \n", len(msgs), cur_time )
        }
    }
}

func store_keeper(writeChan chan WriteMessage, readChan chan ReadMessage){

    store_map := make(map[uint64][]SyncMessage)

    for {
        select {
        case read := <-readChan:
                read.responseChan <- store_map[read.key]
        case write := <-writeChan:
                msg := SyncMessage{
                    message: write.val,
                    t: write.t,
                }

                store_map[write.key] = append( store_map[write.key], msg )
                write.responseChan <- true
        }
    }
}

func writer(mh MessageHandler, writeChan chan WriteMessage){
    for{
        select{
            case m := <-mh.msgChan:

                blob, err := simplejson.NewJson(m.Body)

                if err != nil {
                    log.Fatalf(err.Error())
                }

                val, err := blob.Get("t").String()
                
                if err != nil {
                    log.Fatalf(err.Error())
                }

                msg_time, err := strconv.ParseUint(val, 0, 64)

                if err != nil {
                    log.Fatalf(err.Error())
                }

                mblob, err := blob.MarshalJSON()

                if err != nil {
                    log.Fatalf(err.Error())
                }

                responseChan := make(chan bool)

                msg := WriteMessage{
                    t: msg_time,
                    val: mblob,
                    key: msg_time - msg_time % 40,
                    responseChan: responseChan,
                }

                writeChan <- msg

                if !<-responseChan {
                    log.Fatalf("its broken")
                } else {
                    //log.Println("success")
                }
        }
    }
}

func main(){

    flag.Parse()

    r, err := nsq.NewReader(*topic, *channel)

    if err != nil {
        log.Fatal(err.Error())
    }

    mh := MessageHandler {
        msgChan: make(chan *nsq.Message, 5),
        stopChan: make(chan int),
    }

    wc := make(chan WriteMessage)
    rc := make(chan ReadMessage)

    go store_keeper(wc, rc)
    go writer(mh, wc)
    go emitter(rc, 60 * 1000)

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