package main

import (
    "flag"
    "github.com/bitly/nsq/nsq"
    "github.com/bitly/go-simplejson"
    "time"
    "bufio"
    "net"
    "math/rand"
    "io/ioutil"
    //"strconv"
    "log"
)

var (
    topic            = flag.String("topic", "", "nsq topic")
    channel          = flag.String("channel", "", "nsq topic")
    nsqTCPAddrs      = flag.String("nsqd-tcp-address", "127.0.0.1:4150", "nsqd TCP address")
    nsqHTTPAddrs     = flag.String("nsqd-http-address", "127.0.0.1:4151", "nsqd HTTP address")
    lookupdHTTPAddrs = flag.String("lookupd-http-address", "127.0.0.1:4161", "lookupd HTTP address")
    jsonMsgPath      = flag.String("file","","json file to send")
    timeKey          = flag.String("key","","key that holds time")
)

func writer(tcpAddr string, topic string, msgText []byte, timeKey string){
    
    conn, err := net.DialTimeout("tcp", tcpAddr, time.Second)

    if err != nil {
        panic(err.Error())
    }

    conn.Write(nsq.MagicV2)

    msgJson, _ := simplejson.NewJson(msgText) 

    rw := bufio.NewWriter(bufio.NewWriter(conn))

    c := time.Tick( 5 * time.Second)
    
    r := rand.New(rand.NewSource(99))

    for now := range c {
        count := 0

        batch := make([][]byte, 0)

        for count < 1000 {
            a := int64( r.Float64() * 60000000000 )

            strTime := now.UnixNano() - a
            
            //t := time.Unix(0, strTime)

            //fmt.Println(now.Format("15:04:05") + "-->" + t.Format("15:04:05"))

            msgJson.Set(timeKey, int64(strTime / 1000) )
            b, _ := msgJson.Encode() 

            batch = append( batch, b)
            count++
        }

        log.Println("writing batch")
        cmd, _ := nsq.MultiPublish(topic, batch)
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

    go writer(*nsqTCPAddrs, *topic, content, *timeKey )
    go writer(*nsqTCPAddrs, *topic, content, *timeKey )


    <- stopChan
}