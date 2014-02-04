package blocks

import (
	"encoding/json"
	"github.com/nytlabs/gojee" // jee
	"io/ioutil"
	"log"
	"net/http"
)

// Get, on any inbound message, GETs an external JSON and emits it
func GetHTTP(b *Block) {

	type getRule struct {
		Path string
	}

	var rule *getRule
	var tree *jee.TokenTree
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
			token, err := jee.Lexer(rule.Path)
			if err != nil {
				log.Println(err.Error())
				break
			}
			tree, err = jee.Parser(token)
			if err != nil {
				log.Println(err.Error())
				break
			}
		case msg := <-b.Routes["get_rule"]:
			if rule == nil {
				marshal(msg, &getRule{})
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
			if tree == nil {
				break
			}
			urlInterface, err := jee.Eval(tree, msg.Msg)
			if err != nil {
				log.Println(err.Error())
				break
			}
			urlString, ok := urlInterface.(string)
			if !ok {
				log.Println("couldn't assert url to a string")
				continue
			}

			resp, err := client.Get(urlString)
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
