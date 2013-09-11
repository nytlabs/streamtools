package streamtools

import (
	"github.com/bitly/go-simplejson"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
)

var (
	fail bool = false
)

func getKeyMapping(rules *simplejson.Json) map[string]string {
	out := map[string]string{}
	a, err := rules.Get("keymappings").Array()
	if err != nil {
		log.Fatal(err.Error())
	}
	for _, ai := range a {
		c := ai.(map[string]interface{})
		queryKey := c["queryKey"].(string)
		msgKey := c["msgKey"].(string)
		out[queryKey] = msgKey
	}
	return out
}

func FetchJSON(inChan chan *simplejson.Json, outChan chan *simplejson.Json, RuleChan chan *simplejson.Json) {

	/*
			   key mappings are expected like
		       "keymappings : [
			       {
			           "msgKey": "subject"
			           "queryKey": "glassSubject"
			       },
		           ...
			   ]

			   this will take the 'subject' key from the msg JSON and put its value into the query like &glassSubject=article

	*/
	rules := <-RuleChan
	keymapping := getKeyMapping(rules)
	log.Println("[FETCHJSON] using the folloing as the keymapping:", keymapping)
	endpoint, err := rules.Get("endpoint").String()
	if err != nil {
		log.Fatal(err)
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
					log.Println(keys)
					log.Println(msg)
					log.Fatal(err.Error())
				}
				params.Set(queryKey, value)
			}

			fullUrl := endpoint + params.Encode()
			log.Println("[FETCHJSON] calling", fullUrl)
			resp, err := http.Get(fullUrl)
			if err != nil {
				log.Fatal(err.Error())
			}
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Fatal(err.Error())
			}
			out, err := simplejson.NewJson(body)
			if err != nil {
				log.Fatal(err.Error())
			}
			outChan <- out

		}
	}

}
