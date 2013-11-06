package blocks

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

func PostValue(b *Block) {

	type KeyMapping struct {
		MsgKey   string
		QueryKey string
	}

	type postRule struct {
		Keymapping []KeyMapping
		Endpoint   string
	}

	var rule *postRule
	client := &http.Client{}

	// TODO check the endpoint for happiness
	for {
		select {
		case msg := <-b.AddChan:
			updateOutChans(msg, b)
		case <-b.QuitChan:
			quit(b)
			return
		case msg := <-b.Routes["get_rule"]:
			if rule == nil {
				marshal(msg, &postRule{Keymapping:[]KeyMapping{KeyMapping{}}})
			} else {
				marshal(msg, rule)
			}
		case msg := <-b.Routes["set_rule"]:
			if rule == nil {
				rule = &postRule{}
			}
			unmarshal(msg, rule)

		case msg := <-b.InChan:
			if rule == nil {
				break
			}

			body := make(map[string]interface{})
			for _, keymap := range rule.Keymapping {
				keys := strings.Split(keymap.MsgKey, ".")
				value, err := Get(msg, keys...)
				if err != nil {
					log.Println(err.Error())
				} else {
					Set(body, keymap.QueryKey, value)
				}
			}

			// TODO maybe check the response ?
			postBody, err := json.Marshal(body)
			if err != nil {
				log.Fatal(err.Error())
			}

			// TODO the content-type here is heavily borked but we're using a hack
			req, err := http.NewRequest("POST", rule.Endpoint, bytes.NewReader(postBody))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			req.Close = true

			if err != nil {
			    log.Println(err)
			    break
			}
			
			resp, err := client.Do(req)
			defer resp.Body.Close()

			if err != nil {
			    log.Println(err)
			    break
			}

			bodyBuf := &bytes.Buffer{}
			_, err = bodyBuf.ReadFrom(resp.Body)
		}
	}
}
