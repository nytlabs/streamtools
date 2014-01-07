package blocks

import (
	"github.com/nytlabs/gojee"
	"log"
)

// parse and lex each key
func parseKeys(mapRule map[string]interface{}) (interface{}, error) {
	t := make(map[string]interface{})

	for k, e := range mapRule {
		switch r := e.(type) {
		case map[string]interface{}:
			j, err := parseKeys(r)
			if err != nil {
				return nil, err
			}
			Set(t, k, j)
		case string:
			lexed, err := jee.Lexer(r)
			if err != nil {
				return nil, err
			}
			tree, err := jee.Parser(lexed)
			if err != nil {
				return nil, err
			}
			Set(t, k, tree)
		}
	}

	return t, nil
}

// run jee.eval for each key
func evalMap(mapRule map[string]interface{}, msg map[string]interface{}) (map[string]interface{}, error) {
	nt := make(map[string]interface{})
	for k, _ := range mapRule {
		switch c := mapRule[k].(type) {
		case *jee.TokenTree:
			e, err := jee.Eval(c, msg)
			if err != nil {
				return nil, err
			}
			Set(nt, k, e)
		case map[string]interface{}:
			e, err := evalMap(c, msg)
			if err != nil {
				return nil, err
			}
			msg[k] = Set(nt, k, e)
		}
	}
	return nt, nil
}

// recursively copy map
func recCopy(msg map[string]interface{}) map[string]interface{} {
	n := make(map[string]interface{})

	for k, _ := range msg {
		switch m := msg[k].(type) {
		case map[string]interface{}:
			Set(n, k, recCopy(m))
		default:
			Set(n, k, m)
		}
	}
	return n
}

func Map(b *Block) {
	type maskRule struct {
		Map      interface{}
		Additive bool
	}

	var rule *maskRule
	var parsed interface{}

	for {
		select {
		case m := <-b.Routes["set_rule"]:
			if rule == nil {
				rule = &maskRule{}
			}
			var tmp maskRule
			decode(m, &tmp)

			p, err := parseKeys(tmp.Map.(map[string]interface{}))

			if err == nil {
				parsed = p
				rule = &tmp
			} else {
				log.Println(err)
			}

			marshal(m, rule)
		case r := <-b.Routes["get_rule"]:
			if rule == nil {
				marshal(r, &maskRule{})
			} else {
				marshal(r, rule)
			}
		case msg := <-b.InChan:
			if parsed == nil {
				break
			}

			result := make(map[string]interface{})
			if rule.Additive == true {
				result = recCopy(msg.Msg.(map[string]interface{}))
			}

			in := msg.Msg.(map[string]interface{})
			evaled, err := evalMap(parsed.(map[string]interface{}), in)
			if err != nil {
				log.Println(err)
			}

			for k, _ := range evaled {
				result[k] = evaled[k]
			}

			out := BMsg{
				Msg: result,
			}

			broadcast(b.OutChans, out)
		case msg := <-b.AddChan:
			updateOutChans(msg, b)
		case <-b.QuitChan:
			quit(b)
			return
		}
	}
}
