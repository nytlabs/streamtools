package blocks

import (
	"github.com/nytlabs/gojee"
	"log"
)

func Filter(b *Block) {

	type filterRule struct {
		Filter string
	}

	var rule *filterRule
	var parsed *jee.TokenTree

	for {
		select {
		case msg := <-b.InChan:
			if rule == nil {
				break
			}

			e, err := jee.Eval(parsed, msg.Msg.(map[string]interface{}))
			if err != nil {
				log.Println(err)
				break
			}

			eval, ok := e.(bool)
			if !ok {
				break
			}

			if eval == true {
				out := BMsg{
					Msg: msg.Msg,
				}

				broadcast(b.OutChans, &out)
			}

		case msg := <-b.Routes["set_rule"]:
			if rule == nil {
				rule = &filterRule{}
			}
			var tmp filterRule
			decode(msg, &tmp)

			lexed, err := jee.Lexer(tmp.Filter)
			if err != nil {
				log.Println(err)
				marshal(msg, rule)
				break
			}

			tree, err := jee.Parser(lexed)
			if err != nil {
				log.Println(err)
				marshal(msg, rule)
				break
			}

			rule = &tmp
			parsed = tree

			marshal(msg, rule)

		case msg := <-b.Routes["get_rule"]:
			if rule == nil {
				marshal(msg, &filterRule{})
			} else {
				marshal(msg, rule)
			}
		case msg := <-b.AddChan:
			updateOutChans(msg, b)
		case <-b.QuitChan:
			quit(b)
			return
		}
	}
}
