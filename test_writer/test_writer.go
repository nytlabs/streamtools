package main

import (
    "flag"
    "github.com/bitly/nsq/nsq"
    "github.com/bitly/go-simplejson"
    "time"
    "bufio"
    "net"
    "math/rand"
  //  "fmt"
    "io/ioutil"
    "strconv"
)

var (
    topic            = flag.String("topic", "", "nsq topic")
    channel          = flag.String("channel", "", "nsq topic")
    nsqTCPAddrs      = flag.String("nsqd-tcp-address", "127.0.0.1:4150", "nsqd TCP address")
    nsqHTTPAddrs     = flag.String("nsqd-http-address", "127.0.0.1:4151", "nsqd HTTP address")
    lookupdHTTPAddrs = flag.String("lookupd-http-address", "127.0.0.1:4161", "lookupd HTTP address")
    jsonMsgPath      = flag.String("file","","json file to send")
)

func writer(tcpAddr string, topic string, msgText []byte){
    
    conn, err := net.DialTimeout("tcp", tcpAddr, time.Second)

    if err != nil {
        panic(err.Error())
    }

    conn.Write(nsq.MagicV2)

    msgJson, _ := simplejson.NewJson(msgText) 

    rw := bufio.NewWriter(bufio.NewWriter(conn))

    c := time.Tick( 300 * time.Millisecond)
    
    r := rand.New(rand.NewSource(99))

    for now := range c {

        strTime := now.UnixNano() - int64( r.Float64() * 1000000000 )

        msgJson.Set("t",  strconv.FormatInt( strTime, 10) )
        b, _ := msgJson.Encode() 
        cmd := nsq.Publish(topic, b)
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

    content, err := ioutil.ReadFile(*jsonMsgPath)

    if err != nil {
        //Do something
    }

    go writer(*nsqTCPAddrs, *topic, content )

    <- stopChan
}