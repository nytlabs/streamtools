package blocks

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
)

// Get, on any inbound message, GETs an external JSON and emits it
func GetHTTP(b *Block) {

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
			if err != nil {
				log.Println(err)
				break
			}
			var msg interface{}
			err = json.Unmarshal(body, &msg)
			if err != nil {
				log.Println(err)
				break
			}
			out := BMsg{
				Msg:          msg,
				ResponseChan: nil,
			}
			broadcast(b.OutChans, out)
		}
	}
}
