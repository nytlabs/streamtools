package blocks

import (
	"bytes"
	"github.com/bitly/go-simplejson"
	"log"
	"net/http"
	"strings"
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
		case msg := <-b.InChan:
			body, err := simplejson.NewJson([]byte("{}"))
			if err != nil {
				log.Fatal(err.Error())
			}
			for _, keymap := range rule.Keymapping {
				keys := strings.Split(keymap.MsgKey, ".")
				value, err := msg.GetPath(keys...).String()
				if err != nil {
					log.Fatal(err.Error())
				}
				body.Set(keymap.QueryKey, value)
			}

			// TODO maybe check the response ?
			postBody, err := body.Encode()
			if err != nil {
				log.Fatal(err.Error())
			}
			log.Println(rule.Endpoint, body)
			_, err = http.Post(rule.Endpoint, "application/json", bytes.NewReader(postBody))
			if err != nil {
				log.Fatal(err.Error())
			}
		}
	}
}
