package blocks

import (
	"encoding/json"
	"log"
	"regexp"
)

type opFunc func(interface{}, interface{}) bool

var operators map[string]opFunc

func Filter(b *Block) {

	type filterRule struct {
		Operator   string
		Path       string
		Comparator interface{}
		Invert     bool
	}

	operators = make(map[string]opFunc)

	operators["eq"] = equals
	operators["gt"] = greaterthan
	operators["lt"] = lessthan
	operators["subset"] = subsetof
	operators["regex"] = regexmatch

	var rule *filterRule

	for {
		select {
		case msg := <-b.InChan:
			if rule == nil {
				break
			}

			values := getKeyValues(msg, rule.Path)
			for _, value := range values {
				if operators[rule.Operator](value, rule.Comparator) == !rule.Invert {
					broadcast(b.OutChans, msg)
					break
				}
			}
		case msg := <-b.Routes["set_rule"]:
			if rule == nil {
				rule = &filterRule{}
			}
			// we can't use the standard unmarshal(msg, rule) as we need to make
			// sure the regex compiles, if supplied.
			newRule := &filterRule{}
			err := json.Unmarshal(msg.Msg, &newRule)
			if err != nil {
				log.Println("found errors during unmarshalling")
				log.Println(err.Error())
			}
			if newRule.Operator == "regex" {
				// regex is a bit of a special case
				c, ok := rule.Comparator.(string)
				if !ok {
					log.Println("regex must be a string, not setting rule")
				}
				r, err := regexp.Compile(c)
				if err != nil {
					log.Println("regex did not compile, not setting rule")
					log.Println(err.Error())
				}
				rule = newRule
				rule.Comparator = r
			} else {
				// the simpler rules don't need any futzing
				rule = newRule
			}
			// send the rule back for the response
			m, err := json.Marshal(rule)
			if err != nil {
				log.Println("could not marshal rule")
			}
			msg.ResponseChan <- m
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
