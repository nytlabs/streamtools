package streamtools

import (
	"github.com/bitly/go-simplejson"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

func PollHttp(outChan chan *simplejson.Json, ruleChan chan *simplejson.Json) {
	rules := <-ruleChan
	d, err := rules.Get("duration").Int()
	if err != nil {
		log.Fatal(err.Error())
	}
	endpoint, err := rules.Get("endpoint").String()
	if err != nil {
		log.Fatal(err.Error())
	}

	sampleDuration := time.Duration(d) * time.Second
	ticker := time.NewTicker(sampleDuration)

	for {
		select {
		case <-ruleChan:
		case <-ticker.C:
			log.Println("[POLLHTTP] polling", endpoint)
			resp, err := http.Get(endpoint)
			if err != nil {
				log.Fatal(err.Error())
			}
			defer resp.Body.Close()
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Fatal(err.Error())
			}
			out, err := simplejson.NewJson([]byte(body))
			if err != nil {
				log.Fatal(err.Error())
			}
			outChan <- out
		}
	}
}
