package blocks

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
)

// POSTs an input message to an HTTP endpoint and emits the response
func PostHTTP(b *Block) {

	type postRule struct {
		Endpoint string
	}

	var rule *postRule
	client := &http.Client{}

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
			postBody, err := json.Marshal(msg.Msg)
			if err != nil {
				log.Fatal(err.Error())
				break
			}
			resp, err := client.Post(rule.Endpoint, "application/x-www-form-urlencoded", bytes.NewReader(postBody))
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
			var outMsg interface{}
			err = json.Unmarshal(body, &outMsg)
			if err != nil {
				log.Println(err)
				break
			}
			out := BMsg{
				Msg:          outMsg,
				ResponseChan: nil,
			}
			broadcast(b.OutChans, &out)
		}
	}
}
