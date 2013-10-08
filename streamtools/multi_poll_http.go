package streamtools

import (
	"encoding/json"
	"github.com/bitly/go-simplejson"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

type endpointRule struct {
	Endpoint string
	Name     string
}

func jsonToEndpointRules(endpointsString string) ([]endpointRule, error) {
	var jsonBlob = []byte(endpointsString)
	var rules []endpointRule
	err := json.Unmarshal(jsonBlob, &rules)
	return rules, err
}

func MultiPollHttp(outChan chan *simplejson.Json, ruleChan chan *simplejson.Json) {
	rules := <-ruleChan
	d, err := rules.Get("duration").Int()
	if err != nil {
		log.Fatal(err.Error())
	}
	endpointsString, err := rules.Get("endpoints").String()
	log.Println("using", endpointsString)
	if err != nil {
		log.Fatal(err.Error())
	}

	// unpack the endpoints
	endpointsSlice, err := jsonToEndpointRules(endpointsString)
	if err != nil {
		log.Fatal(err.Error())
	}

	sampleDuration := time.Duration(d) * time.Second
	ticker := time.NewTicker(sampleDuration)
	blobs := make([]*simplejson.Json, len(endpointsSlice))

	for {
		select {
		case <-ruleChan:
		case <-ticker.C:
			out, err := simplejson.NewJson([]byte("{}"))
			for i, rule := range endpointsSlice {
				log.Println("polling", rule.Endpoint)
				resp, err := http.Get(rule.Endpoint)
				if err != nil {
					log.Fatal(err.Error())
				}
				defer resp.Body.Close()
				body, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					log.Fatal(err.Error())
				}
				outBlob, err := simplejson.NewJson([]byte("{}"))
				data, err := simplejson.NewJson(body)
				outBlob.Set("name", rule.Name)
				outBlob.Set("data", data)
				log.Println(outBlob)
				blobs[i] = outBlob
			}
			if err != nil {
				log.Fatal(err.Error())
			}
			out.Set("results", blobs)
			outChan <- out
		}
	}
}
