package streamtools

import (
	"github.com/bitly/go-simplejson"
	"log"
)

func maskJSON(mask *simplejson.Json, input *simplejson.Json) *simplejson.Json {
	t, _ := simplejson.NewJson([]byte(`{}`))

	maskMap, err := mask.Map()
	if err != nil {
		log.Println(err.Error())
	}

	inputMap, err := input.Map()
	if err != nil {
		log.Println(err.Error())
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

func Mask(inChan chan *simplejson.Json, outChan chan *simplejson.Json, RuleChan chan *simplejson.Json) {
	mask, _ := simplejson.NewJson([]byte(`{}`))
	for {
		select {
		case inputRule := <-RuleChan:
			mask = inputRule
		case msg := <-inChan:
			outChan <- maskJSON(mask, msg)
		}
	}
}
