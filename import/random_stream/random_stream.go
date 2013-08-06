// Fakes a stream of data at your convenience. Doesn't yet quite generate random stream.
// Takes a string as input and puts a random-time-offset message on an nsq.

package main

import (
	"flag"
	"github.com/bitly/go-simplejson"
	"io/ioutil"
	"math/rand"
	"time"
	//"strconv"
	"bytes"
	"log"
	"net/http"
)

var (
	topic        = flag.String("topic", "random", "nsq topic")
	jsonMsgPath  = flag.String("file", "test.json", "json file to send")
	timeKey      = flag.String("key", "t", "key that holds time")

	nsqHTTPAddrs = "127.0.0.1:4151"
)

func writer(msgText []byte) {
	msgJson, _ := simplejson.NewJson(msgText)
	client := &http.Client{}

	c := time.Tick(5 * time.Second)
	r := rand.New(rand.NewSource(99))

	for now := range c {
		a := int64(r.Float64() * 10000000000)
		strTime := now.UnixNano() - a
		msgJson.Set(*timeKey, int64(strTime/1000000))
		outMsg, _ := msgJson.Encode()
		msgReader := bytes.NewReader(outMsg)
		resp, err := client.Post("http://"+nsqHTTPAddrs+"/put?topic="+*topic, "data/multi-part", msgReader)
		if err != nil {
			log.Fatalf(err.Error())
		}
		resp.Body.Close()
	}
}

func main() {

	flag.Parse()

	stopChan := make(chan int)

	content, err := ioutil.ReadFile(*jsonMsgPath)

	if err != nil {
		log.Fatal(err.Error())
	}

	go writer(content)

	<-stopChan
}
