package streamtools

import (
	"github.com/bitly/go-simplejson"
	"log"
)

func maskJSON(mask *simplejson.Json, input *simplejson.Json) *simplejson.Json {
	t, _ := simplejson.NewJson([]byte(`{}`))

	maskMap, err := mask.Map()
	if err != nil {
		log.Fatal(err.Error())
	}

	if len(maskMap) == 0 {
		return input
	}

	inputMap, err := input.Map()
	if err != nil {
		log.Fatal(err.Error())
	}

	for k, _ := range maskMap {
		switch inputMap[k].(type) {
		case map[string]interface{}:
			t.Set(k, maskJSON(mask.Get(k), input.Get(k)))
		default:
			t.Set(k, input.Get(k))
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
//	{"a":{}, "b":{"d":{}},"x":{}}
// And an incoming message looks like:
//	{"a":24, "b":{"c":"test", "d":[1,3,4]}, "f":5, "x":{"y":5, "z":10}}
// The resulting object after the application of Mask would be:
//	{"a":24, "b":{"d":[1,3,4]}, "x":{"y":5, "z":10}
func Mask(inChan chan *simplejson.Json, outChan chan *simplejson.Json, RuleChan chan *simplejson.Json) {
	mask := <-RuleChan
	for {
		select {
		case inputRule := <-RuleChan:
			mask = inputRule
		case msg := <-inChan:
			outChan <- maskJSON(mask, msg)
		}
	}
}
