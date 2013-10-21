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
			body := simplejson.New()
			for _, keymap := range rule.Keymapping {
				keys := strings.Split(keymap.MsgKey, ".")
				value := msg.GetPath(keys...).Interface()
				body.Set(keymap.QueryKey, value)

			}

			// TODO maybe check the response ?
			postBody, err := body.Encode()
			if err != nil {
				log.Fatal(err.Error())
			}
			log.Println(body)

			// TODO the content-type here is heavily borked but we're using a hack
			http.Post(rule.Endpoint, "application/x-www-form-urlencoded", bytes.NewReader(postBody))
		}
	}
}
