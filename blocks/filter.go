package blocks

import (
	"log"
)

type opFunc func(interface{}, interface{}) bool

var operators map[string]opFunc

func Filter(b *Block) {

	type filterRule struct {
		Operator   string
		Path       string
		Comparator interface{}
	}

	operators = make(map[string]opFunc)

	operators["eq"] = equals
	operators["gt"] = greaterthan
	operators["lt"] = lessthan
	operators["subset"] = subsetof

	rule := &filterRule{}
	unmarshal(<-b.Routes["set_rule"], &rule)

	for {
		select {
		case msg := <-b.InChan:
			values := getKeyValues(msg.Interface(), rule.Path)
			log.Println("VALUES", values)
			for _, value := range values {
				log.Println(value)
				log.Println(rule.Operator)
				log.Println(operators[rule.Operator](value, rule.Comparator))
				if operators[rule.Operator](value, rule.Comparator) {
					broadcast(b.OutChans, msg)
					break
				}
			}
		case msg := <-b.AddChan:
			updateOutChans(msg, b)
		}
	}

}
