package streamtools

import (
	"github.com/bitly/go-simplejson"
	"log"
	"net/http"
	"strings"
)

func FromHTTP(outChan chan *simplejson.Json, ruleChan chan *simplejson.Json) {

	rule := <-ruleChan

	url, err := rule.Get("endpoint").String()
	if err != nil {
		log.Fatal(err)
	}
	auth, err := rule.Get("auth").String()
	if err != nil {
		log.Fatal(err)
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal(err)
	}
	if len(auth) > 0 {
		req.SetBasicAuth(strings.Split(auth, ":")[0], strings.Split(auth, ":")[1])
	}
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()

	for {
		select {
		case <-ruleChan:

		default:
			body := make([]byte, 1024)
			_, err := res.Body.Read(body)
			if err != nil {
				log.Fatal(err)
			}
			// TODO worry about better detection of end of body
			idx := strings.LastIndex(string(body), "}")
			if idx != -1 {
				msg, err := simplejson.NewJson(body[:idx+1])
				if err != nil {
					log.Fatal(err.Error())
				}
				outChan <- msg
			}
		}
	}

}
