package main

import (
    "flag"
    "github.com/bitly/go-simplejson"
    "time"
    "math/rand"
    "io/ioutil"
    //"strconv"
    "log"
    "net/http"
    "bytes"
)

var (
    topic            = flag.String("topic", "test", "nsq topic")
    nsqHTTPAddrs     = flag.String("nsqd-tcp-address", "127.0.0.1:4151", "nsqd TCP address")
    jsonMsgPath      = flag.String("file","test.json","json file to send")
    timeKey          = flag.String("key","t","key that holds time")
)

func writer(msgText []byte){
    msgJson, _ := simplejson.NewJson(msgText) 
    client := &http.Client{}

    c := time.Tick( 5 * time.Second)
    r := rand.New(rand.NewSource(99))

    for now := range c {
        a := int64( r.Float64() * 10000000000 )
        strTime := now.UnixNano() - a
        msgJson.Set(*timeKey, int64(strTime / 1000000) )
        outMsg, _ := msgJson.Encode() 
        msgReader := bytes.NewReader(outMsg)
        resp, err := client.Post("http://" + *nsqHTTPAddrs + "/put?topic=" + *topic,"data/multi-part", msgReader)
        if err != nil {
            log.Fatalf(err.Error())
        }
        resp.Body.Close()
    }
}

func main(){
    
    flag.Parse()

    stopChan := make(chan int)

    content, err := ioutil.ReadFile(*jsonMsgPath)

    if err != nil {
        log.Fatal(err.Error())
    }

    go writer(content)

    <- stopChan
}