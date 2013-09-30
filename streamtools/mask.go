package streamtools

import (
	"github.com/bitly/go-simplejson"
	"log"
)

func maskJSON(mask *simplejson.Json, input *simplejson.Json) *simplejson.Json {
	t, _ := simplejson.NewJson([]byte(`{}`))

	log.Println("mask 3", mask)

	maskMap, err := mask.Map()
	if err != nil {
		log.Println("mask error", mask)
		log.Println(err.Error())
	}

	inputMap, err := input.Map()
	if err != nil {
		log.Println(err.Error())
	}

	for k, _ := range maskMap {
		switch inputMap[k].(type) {
		case map[string]interface{}:
			log.Println("mask 4", mask, 'k', k)
			log.Println(mask.Get(k), input.Get(k))
			t.Set(k, maskJSON(mask.Get(k), input.Get(k)))
		default:
			t.Set(k, input.Get(k))
		}
	}
	return t
}

func Mask(inChan chan *simplejson.Json, outChan chan *simplejson.Json, RuleChan chan *simplejson.Json) {
	mask := <-RuleChan
	log.Println("mask 1", mask)
	for {
		select {
		case inputRule := <-RuleChan:
			log.Println("argh")
			mask = inputRule
		case msg := <-inChan:
			log.Println("mask 2", mask)
			outChan <- maskJSON(mask, msg)
		}
	}
}
