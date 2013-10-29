package blocks

import (
	"bytes"
	"log"
	"net/http"
	"strings"
	"encoding/json"
)

func Post(b *Block) {

	type KeyMapping struct {
		MsgKey   string
		QueryKey string
	}

	type postRule struct {
		Keymapping []KeyMapping
		Endpoint   string
	}

	rule := &postRule{}

	unmarshal(<-b.Routes["set_rule"], &rule)

	// TODO check the endpoint for happiness
	for {
		select {
		case msg := <-b.AddChan:
			updateOutChans(msg, b)
		case <-b.QuitChan:
			quit(b)
			return
		case msg := <-b.InChan:
			var body BMsg
			for _, keymap := range rule.Keymapping {
				keys := strings.Split(keymap.MsgKey, ".")
				value, err := Get(msg, keys...)
				if err != nil{
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
			log.Println(body)

			// TODO the content-type here is heavily borked but we're using a hack
			http.Post(rule.Endpoint, "application/x-www-form-urlencoded", bytes.NewReader(postBody))
		}
	}
}
