package blocks

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
)

func Post(b *Block) {

	type postRule struct {
		Endpoint string
	}

	var rule *postRule

	// TODO check the endpoint for happiness
	for {
		select {
		case msg := <-b.AddChan:
			updateOutChans(msg, b)
		case msg := <-b.Routes["set_rule"]:
			if rule == nil {
				rule = &postRule{}
			}
			unmarshal(msg, rule)

		case msg := <-b.Routes["get_rule"]:
			if rule == nil {
				marshal(msg, &postRule{})
			} else {
				marshal(msg, rule)
			}
		case <-b.QuitChan:
			quit(b)
			return
		case msg := <-b.InChan:
			if rule == nil {
				break
			}
			// TODO maybe check the response ?
			postBody, err := json.Marshal(msg)
			if err != nil {
				log.Fatal(err.Error())
			}

			// TODO the content-type here is heavily borked but we're using a hack
			resp, err := http.Post(rule.Endpoint, "application/x-www-form-urlencoded", bytes.NewReader(postBody))
			if err != nil {
				log.Println(err.Error())
			} else {
				defer resp.Body.Close()
			}
		}
	}
}
