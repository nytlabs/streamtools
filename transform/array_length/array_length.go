package main

import (
    "flag"
    "log"
    "github.com/bitly/nsq/nsq"
    "github.com/bitly/go-simplejson"
)

var (
    inTopic = flag.String("in_topic", "", "topic to read from")
    outTopic = flag.String("out_topic", "", "topic to write to")
    arrayKey = flag.String("key", "", "key of the array whose length you would like")
    lookupdHTTPAddrs = "127.0.0.1:4161"
    nsqdAddr = "127.0.0.1:4150"
)

type SyncHandler struct {
    msgChan chan *nsq.Message
}

func (self *SyncHandler) HandleMessage(m *nsq.Message) error {
    self.msgChan <- m
    return nil
}

func GetLength (msgChan chan *nsq.Message, outChan chan int) {
    for {
        select {
            case m := <- msgChan:
                blob, err := simplejson.NewJson(m.Body)
                if err != nil {
                    log.Fatal(err.Error())
                }
                arr, err := blob.Get(*arrayKey).Array()
                if err != nil {
                    log.Fatal(err.Error())
                }
                l := len(arr)
                log.Println(l)
                outChan <- l
        }
    }
}

func Writer (outChan chan int) {
    w := nsq.NewWriter(0)
    err := w.ConnectToNSQ(nsqdAddr)
    if err != nil {
        log.Fatal(err.Error())
    }
    for {
        select {
        case l := <- outChan:
            msg := []byte("{\"len_"+ *arrayKey +"\":" + string(l) + "}")
            frameType, data, err := w.Publish(*outTopic, msg)
            if err != nil {
                log.Fatalf("frametype %d data %s error %s", frameType, string(data), err.Error())
            }
        }
    }

}

func main() {

    flag.Parse()
    channel := "length:"+*arrayKey
    r, _ := nsq.NewReader(*inTopic, channel)

    msgChan := make(chan *nsq.Message)
    outChan := make(chan int)

    go GetLength(msgChan, outChan)
    go Writer(outChan)

    sh := SyncHandler{
       msgChan: msgChan,
    }
    r.AddHandler(&sh)
}

