package main

import (
    "flag"
    "github.com/bitly/nsq/nsq"
    "time"
    "bufio"
    "net"
)

var (
    topic            = flag.String("topic", "", "nsq topic")
    channel          = flag.String("channel", "", "nsq topic")
    nsqTCPAddrs      = flag.String("nsqd-tcp-address", "127.0.0.1:4150", "nsqd TCP address")
    nsqHTTPAddrs     = flag.String("nsqd-http-address", "127.0.0.1:4151", "nsqd HTTP address")
    lookupdHTTPAddrs = flag.String("lookupd-http-address", "127.0.0.1:4161", "lookupd HTTP address")
)

func writer(tcpAddr string, topic string){
    
    conn, err := net.DialTimeout("tcp", tcpAddr, time.Second)

    if err != nil {
        panic(err.Error())
    }
    conn.Write(nsq.MagicV2)

    rw := bufio.NewWriter(bufio.NewWriter(conn))

    for {
        msg := []byte("hello")
        cmd := nsq.Publish(topic, msg)
        err := cmd.Write(rw)
        if err != nil {
            panic(err.Error())
        }
        err = rw.Flush()
    }
}

func main(){
    
    flag.Parse()

    stopChan := make(chan int)

    go writer(*nsqTCPAddrs, *topic)

    <- stopChan
}