package streamtools

import (
	"bytes"
	"github.com/bitly/go-simplejson"
	"log"
	"net/http"
	"strings"
)

func PostHTTP(inChan chan *simplejson.Json, RuleChan chan *simplejson.Json) {

	rules := <-RuleChan
	keymapping := getKeyMapping(rules)
	log.Println("[HTTPPOST] using the folloing as the keymapping:", keymapping)
	endpoint, err := rules.Get("endpoint").String()
	if err != nil {
		log.Fatal(err.Error())
	}
	// TODO check the endpoint for happiness
	for {
		select {
		case <-RuleChan:
		case msg := <-inChan:
			body, err := simplejson.NewJson([]byte("{}"))
			if err != nil {
				log.Fatal(err.Error())
			}
			for queryKey, msgKey := range keymapping {
				keys := strings.Split(msgKey, ".")
				value, err := msg.GetPath(keys...).String()
				if err != nil {
					log.Fatal(err.Error())
				}
				body.Set(queryKey, value)
			}

			// TODO maybe check the response ?
			postBody, err := body.Encode()
			if err != nil {
				log.Fatal(err.Error())
			}
			_, err = http.Post(endpoint, "application/json", bytes.NewReader(postBody))
			if err != nil {
				log.Fatal(err.Error())
			}
		}
	}

}
