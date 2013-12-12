package blocks

import (
	"log"
	"regexp"
	"github.com/mitchellh/mapstructure"
)

type opFunc func(interface{}, interface{}) bool

var operators map[string]opFunc

// Filter queries a message for all values that match the given Path parameter
// and compares those values to a rule given an operator and a comparator. If
// any value passes the filter operation then the message is broadcast. If no
// value satisfies the filter operation the message is ignored.
//
// Filter is capable of traversing arrays that contain elements with and
// without keys. For example, if Path is set to
// 		foo.bar[]
// All elements within the "bar" array will be compared to the filter operation.
// In the case of
//		foo.bar[].baz
// Only the value of the "baz" keys within elements of the "bar" array will be
// used for the filter operation.
//
// There are four filter comparators: greater than "gt", less than "lt", equal
// to "eq" and subset of "subset".
//
// gt, lt, eq operations are available for values of a number type.
// eq, subset operations are avilable for values of a string type.
// eq operations are availble for value of a bool or null type.
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
	operators["keyin"] = keyin

	var rule *filterRule

	for {
		select {
		case msg := <-b.InChan:
			if rule == nil {
				break
			}

			values := getKeyValues(msg, rule.Path)
			for _, value := range values {
				op, ok := operators[rule.Operator]
				if !ok {
					log.Println("specified operator does not exist")
					break
				}
				
				if op(value, rule.Comparator) == !rule.Invert {
					broadcast(b.OutChans, msg)
					break
				}
			}

			if len(values) == 0 && rule.Invert {
				broadcast(b.OutChans, msg)
			}

		case msg := <-b.Routes["set_rule"]:
			if rule == nil {
				rule = &filterRule{}
			}
			// this is used to send back on the response chan
			var ruleRegexString string
			// we can't use the standard unmarshal(msg, rule) as we need to make
			// sure the regex compiles, if supplied.

			var inMsg BMsg
			rr, isRouteResponse := msg.(RouteResponse)
			if isRouteResponse {
				inMsg = rr.Msg
			} else {
				inMsg = msg
			}

			newRule := &filterRule{}
			err := mapstructure.Decode(inMsg, newRule)
			if err != nil {
				log.Println("found errors during decoding")
				log.Println(err.Error())
				if isRouteResponse {
					rr.ResponseChan <- rule
				}
				break
			}
			if _, ok := operators[newRule.Operator]; !ok {
				log.Println("invalid operator")
				if isRouteResponse {
					rr.ResponseChan <- rule
				}
				break
			}
			if newRule.Operator == "regex" {
				// regex is a bit of a special case
				c, ok := newRule.Comparator.(string)
				if !ok {
					log.Println("regex must be a string, not setting rule")
					if isRouteResponse {
						rr.ResponseChan <- rule
					}
					break
				}
				r, err := regexp.Compile(c)
				if err != nil {
					log.Println("regex did not compile, not setting rule")
					log.Println(err.Error())
					if isRouteResponse {
						rr.ResponseChan <- rule
					}
					break
				}
				rule = newRule
				rule.Comparator = r
				ruleRegexString = r.String()
			} else {
				// the simpler rules don't need any futzing
				rule = newRule
			}
			// send the rule back for the response
			out_rule := *rule
			if rule.Operator == "regex" {
				// we replace the regex in the outgoing rule with its string
				// representation so we can marshal it correctly.
				out_rule.Comparator = ruleRegexString
			}
			
			if isRouteResponse{
				rr.ResponseChan <- out_rule
			}

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
