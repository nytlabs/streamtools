package main

import (
    "flag"
    "log"
    "github.com/bitly/nsq/nsq"
)

var (
    topic            = flag.String("topic", "", "nsq topic")
    channel          = flag.String("channel", "", "nsq topic")
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

func writer(mh MessageHandler) {
    for {
        select {
            case msg := <-mh.msgChan:
                log.Println( string( msg.Body ) )
        }
    }
}

func main(){
    
    flag.Parse()

    stopChan := make(chan int)

    r, err := nsq.NewReader(*topic, *channel)

    mh := MessageHandler{
        msgChan: make(chan *nsq.Message, 5),
    }

    go writer(mh)

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