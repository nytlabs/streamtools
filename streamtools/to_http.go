package streamtools

import (
	"github.com/bitly/go-simplejson"
	"log"
	"net/http"
	"net/url"
	"strings"
)

func ToHTTP(inChan chan *simplejson.Json, RuleChan chan *simplejson.Json) {

	rules := <-RuleChan
	keymapping := getKeyMapping(rules)
	log.Println("using the folloing as the keymapping:", keymapping)
	endpoint, err := rules.Get("endpoint").String()
	if err != nil {
		log.Fatal(err.Error())
	}
	// TODO check the endpoint for happiness
	if !strings.HasSuffix(endpoint, "?") {
		endpoint = endpoint + "?"
	}
	for {
		select {
		case <-RuleChan:
		case msg := <-inChan:
			params := url.Values{}
			for queryKey, msgKey := range keymapping {
				keys := strings.Split(msgKey, ".")
				value, err := msg.GetPath(keys...).String()
				if err != nil {
					log.Println(err.Error())
				}
				params.Set(queryKey, value)
			}

			fullUrl := endpoint + params.Encode()
			log.Println("calling", fullUrl)
			// TODO maybe check the response ?
			_, err := http.Get(fullUrl)
			if err != nil {
				log.Fatal(err.Error())
			}
		}
	}

}
