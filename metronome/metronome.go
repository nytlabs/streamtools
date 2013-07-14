package main

import (
    "time"
    "flag"
    "log"
    "github.com/bitly/nsq/nsq"
    "bytes"
)

var (
    waitTime         = flag.Int("wait", 1000, "wait time in ms")
    topic            = flag.String("topic", "monitor", "nsq topic")
    channel          = flag.String("channel", "monitorreader", "nsq topic")
    maxInFlight      = flag.Int("max-in-flight", 1000, "max number of messages to allow in flight")
    nsqTCPAddrs      = flag.String("nsqd-tcp-address", "127.0.0.1:4150", "nsqd TCP address")
    nsqHTTPAddrs     = flag.String("nsqd-http-address", "127.0.0.1:4151", "nsqd HTTP address")
    lookupdHTTPAddrs = flag.String("lookupd-http-address", "127.0.0.1:4161", "lookupd HTTP address")
)

type MessageHandler struct {
    msgChan  chan *nsq.Message
}

func (self *MessageHandler) HandleMessage(message *nsq.Message, responseChannel chan *nsq.FinishedMessage) {
    self.msgChan <- message
    responseChannel <- &nsq.FinishedMessage{message.Id, 0, true}
}

func writer(mh MessageHandler, writeChan chan []byte) {
    for {
        select {
            case msg := <-mh.msgChan:
                writeChan <- msg.Body
        }
    }
}

func dumper(dumpChan chan int, waitTime int){
    c := time.Tick( time.Duration(waitTime) * time.Millisecond)
    for _ = range c {
        dumpChan <- 1
    }
}

func bucket(msgChan chan []byte, dumpChan chan int){
    msgs := []byte{}
    for {
        select {
            case msg := <-msgChan:
                msgs = bytes.Join( [][]byte{msgs, msg}, []byte{'\n'} )
            case _ = <-dumpChan: 
                log.Println( string(msgs) )
                msgs = []byte{}
        }
    }
}

func main(){
    
    flag.Parse()

    stopChan := make(chan int)
    dumpChan := make(chan int)
    msgChan := make(chan []byte)

    r, err := nsq.NewReader(*topic, *channel)

    mh := MessageHandler{
        msgChan: make(chan *nsq.Message, 5),
    }

    go bucket(msgChan, dumpChan)
    go dumper(dumpChan, *waitTime)
    go writer(mh, msgChan)

    r.AddAsyncHandler(&mh)

    err = r.ConnectToNSQ(*nsqTCPAddrs)
    if err != nil {
        log.Fatalf(err.Error())
    }
    err = r.ConnectToLookupd(*lookupdHTTPAddrs)
    if err != nil {
        log.Fatalf(err.Error())
    }

    <- stopChan
}