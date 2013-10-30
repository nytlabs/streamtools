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
