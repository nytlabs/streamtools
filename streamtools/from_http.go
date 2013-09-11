package streamtools

import (
	"bytes"
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

	var body bytes.Buffer
	crlf := []byte{125, 10}

	for {
		select {
		case <-ruleChan:

		default:
			buffer := make([]byte, 5*1024)
			p, err := res.Body.Read(buffer)
			if err != nil {
				log.Fatal(err.Error())
			}
			body.Write(buffer)
			if bytes.Equal(crlf, buffer[p-2:p]) {
				msg, err := simplejson.NewJson(body.Bytes())
				if err != nil {
					log.Fatal(err.Error())
				}
				outChan <- msg
				body.Reset()
			}
		}
	}

}
