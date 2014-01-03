package blocks

import "github.com/nytlabs/gojee"
import "fmt"

/*func maskJSON(maskMap map[string]interface{}, input map[string]interface{}) interface{} {
	t := make(map[string]interface{})

	if len(maskMap) == 0 {
		return input
	}

	for k, _ := range maskMap {
		val, ok := input[k]
		if ok {
			switch v := val.(type) {
			case map[string]interface{}:
				maskNext, ok := maskMap[k].(map[string]interface{})
				if ok {
					Set(t, k, maskJSON(maskNext, v))
				} else {
					Set(t, k, v)
				}
			default:
				Set(t, k, val)
			}
		}
	}

	return t
}*/


func parseKeys(mapRule map[string]interface{}) (interface{}, error) {
	t := make(map[string]interface{})

	for k, e := range mapRule {
		switch r := e.(type){
		case map[string]interface{}:
			j, err := parseKeys(r)
			if err != nil {
				return nil, err
			}
			Set(t,k,j)
		case string:
			lexed, err := jee.Lexer(r)
			if err != nil{
				return nil, err
			}
			tree, err := jee.Parser(lexed)
			if err != nil {
				return nil, err
			}
			Set(t,k, tree)
		}
	}

	return t, nil
}

func evalMap(mapRule map[string]interface{}, msg map[string]interface{}) (map[string]interface{}, error){
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

// Mask modifies a JSON stream with an additive key filter. Mask uses the JSON
// object recieved through the rule channel to determine which keys should be
// included in the resulting object. An empty JSON object ({}) is used as the
// notation to include all values for a key.
//
// For instance, if the JSON rule is:
//        {"a":{}, "b":{"d":{}},"x":{}}
// And an incoming message looks like:
//        {"a":24, "b":{"c":"test", "d":[1,3,4]}, "f":5, "x":{"y":5, "z":10}}
// The resulting object after the application of Mask would be:
//        {"a":24, "b":{"d":[1,3,4]}, "x":{"y":5, "z":10}}
func Map(b *Block) {
	type maskRule struct {
		Map interface{}
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
				fmt.Println(err)
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

			in := msg.Msg.(map[string]interface{})
			//var result 
			result, err := evalMap(parsed.(map[string]interface{}), in)
			if err != nil {
				fmt.Println(err)
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
