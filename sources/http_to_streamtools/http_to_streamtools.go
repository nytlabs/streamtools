package main

import (
	"flag"
    "github.com/bitly/nsq/nsq"
    "log"
    "net/http"
    "os"
    "strings"
)

var (
	url   = flag.String("url", "", "streaming url endpoint")
	topic = flag.String("topic", "", "topic to write to")

    addr = "127.0.0.1:4150"
)

func writer(outChan *chan []byte, w *nsq.Writer) {
    for {
        select {
        case msg := <- *outChan:
            frameType, data, err := w.Publish(*topic, msg)
            if err != nil {
                log.Fatalf("frametype %d data %s error %s", frameType, string(data), err.Error())
            }
        }
    }
}

func reader(outChan *chan []byte) {
    res, err := http.Get(*url)
    if err != nil {
        log.Fatal(err)
    }
    defer res.Body.Close()
    for {
        // TODO worry about what happens when the message is bigger than 1024
        body := make([]byte, 1024)
        _, err := res.Body.Read(body)
        if err != nil {
            log.Fatal(err)
        }
        // TODO worry about better detection of end of body
        idx := strings.LastIndex(string(body), "}")
        if idx != -1 {
            msg := body[:idx+1]
            *outChan <- msg
            log.Println(string(msg))
        }
    }
}

func main() {
	flag.Parse()
    outChan := make(chan []byte)
    sigChan := make(chan os.Signal, 1)

    w := nsq.NewWriter(0)
    err := w.ConnectToNSQ(addr)
    if err != nil {
        log.Fatal(err.Error())
    }
    go writer(&outChan, w)
    go reader(&outChan)

    <-sigChan

}
