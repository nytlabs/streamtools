package blocks

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
)

// Get GETs an external JSON and emits it
func HTTPGet(b *Block) {

	type getRule struct {
		Endpoint string
	}

	var rule *getRule
	client := &http.Client{}

	// TODO check the endpoint for happiness
	for {
		select {
		case msg := <-b.AddChan:
			updateOutChans(msg, b)
		case msg := <-b.Routes["set_rule"]:
			if rule == nil {
				rule = &getRule{}
			}
			unmarshal(msg, rule)

		case msg := <-b.Routes["get_rule"]:
			if rule == nil {
				marshal(msg, &getRule{})
			} else {
				marshal(msg, rule)
			}
		case <-b.QuitChan:
			quit(b)
			return
		case <-b.InChan:
			if rule == nil {
				break
			}

			resp, err := client.Get(rule.Endpoint)
			defer resp.Body.Close()
			if err != nil {
				log.Println(err.Error())
				break
			}

			body, err := ioutil.ReadAll(resp.Body)
			var msg BMsg
			err = json.Unmarshal(body, &msg)
			if err != nil {
				log.Println(err)
				break
			}
			broadcast(b.OutChans, msg)
		}
	}
}
