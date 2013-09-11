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
	// these are the possible delimiters
	d1 := []byte{125, 10} // this is }\n
	d2 := []byte{13, 10}  // this is CRLF

	for {
		select {
		case <-ruleChan:

		default:
			buffer := make([]byte, 5*1024)
			p, err := res.Body.Read(buffer)
			if err != nil {
				log.Fatal(err.Error())
			}
			body.Write(buffer[:p])
			if bytes.Equal(d1, buffer[p-2:p]) { // ended with }\n
				for _, blob := range bytes.Split(body.Bytes(), []byte{10}) { // split on new line in case there are multuple messages per buffer
					if len(blob) > 0 {
						msg, err := simplejson.NewJson(blob)
						if err != nil {
							log.Fatal(err.Error())
						}
						outChan <- msg
					}
				}
				body.Reset()
			} else if bytes.Equal(d2, buffer[p-2:p]) { // ended with CRLF which we don't care about
				body.Reset()
			}
		}
	}

}
