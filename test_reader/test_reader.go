package main

import (
    "flag"
    "log"
    "github.com/bitly/nsq/nsq"
    "github.com/bitly/go-simplejson"
    "time"
)

var (
    topic            = flag.String("topic", "", "nsq topic")
    channel          = flag.String("channel", "", "nsq topic")
    nsqTCPAddrs      = flag.String("nsqd-tcp-address", "127.0.0.1:4150", "nsqd TCP address")
    nsqHTTPAddrs     = flag.String("nsqd-http-address", "127.0.0.1:4151", "nsqd HTTP address")
    lookupdHTTPAddrs = flag.String("lookupd-http-address", "127.0.0.1:4161", "lookupd HTTP address")
    timeKey          = flag.String("key","","key that holds time")
)

type MessageHandler struct {
    msgChan  chan *nsq.Message
}

func (self *MessageHandler) HandleMessage(message *nsq.Message, responseChannel chan *nsq.FinishedMessage) {
    self.msgChan <- message
    responseChannel <- &nsq.FinishedMessage{message.Id, 0, true}
}

func writer(mh MessageHandler, timeKey string) {
    const layout = "15:04:05"

    for {
        select {
            case msg := <-mh.msgChan:
                blob, err := simplejson.NewJson(msg.Body)
                if err == nil {
                    //log.Println(string(msg.Body))
                    //log.Fatalf(err.Error())
                    msg_time, _ := blob.Get(timeKey).Int64()

                    t := time.Unix(0, msg_time * 1000 * 1000)

                    log.Println(t.Format(layout) )
                }
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

    go writer(mh, *timeKey)

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