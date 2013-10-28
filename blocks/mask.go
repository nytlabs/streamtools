package blocks

import (
	"github.com/bitly/go-simplejson"
	"log"
)

func maskJSON(maskMap map[string]interface{}, input *simplejson.Json) *simplejson.Json {
        t, _ := simplejson.NewJson([]byte(`{}`))

        if len(maskMap) == 0 {
                return input
        }

        inputMap, err := input.Map()
        if err != nil {
        	log.Fatal(err.Error())
        }

        for k, _ := range maskMap {
        	val, _ := inputMap[k]
        	switch val.(type){
        	case *simplejson.Json:
	        	valJ := val.(*simplejson.Json)
	        	_, err := valJ.Map()
	        	if err != nil {
	        		log.Println("could not convert ", inputMap[k] )
	        		t.Set(k, inputMap[k])
	        	} else {
	        		t.Set(k, maskJSON(maskMap[k].(map[string]interface{}), valJ ))
	        	}
	        default:
	        	t.Set(k, val)
	        }

        }

        return t
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
func Mask(b *Block) {
	type maskRule struct {
		Mask interface{}
	}
	rule := &maskRule{}

	unmarshal(<-b.Routes["set_rule"], &rule)

	for {
		select {
		case m := <-b.Routes["set_rule"]:
			unmarshal(m, &rule)
		case r := <-b.Routes["get_rule"]:
			marshal(r, rule)
		case msg := <-b.InChan:
			/*m, err := msg.Map()
			if err != nil {
				log.Println(err.Error())
			}*/
			broadcast(b.OutChans, maskJSON(rule.Mask.(map[string]interface{}), msg))
		case msg := <-b.AddChan:
			updateOutChans(msg, b)
		case <-b.QuitChan:
			quit(b)
			return
		}
	}
}
