package blocks

func maskJSON(maskMap map[string]interface{}, input map[string]interface{}) interface{} {
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
	var rule *maskRule

	for {
		select {
		case m := <-b.Routes["set_rule"]:
			if rule == nil {
				rule = &maskRule{}
			}
			unmarshal(m, rule)
		case r := <-b.Routes["get_rule"]:
			if rule == nil {
				marshal(r, &maskRule{})
			} else {
				marshal(r, rule)
			}
		case msg := <-b.InChan:
			if rule == nil {
				break
			}

			msgMap, msgOk := msg.(map[string]interface{})
			maskMap, maskOk := rule.Mask.(map[string]interface{})
			if msgOk && maskOk {
				broadcast(b.OutChans, maskJSON(maskMap, msgMap))
			}
		case msg := <-b.AddChan:
			updateOutChans(msg, b)
		case <-b.QuitChan:
			quit(b)
			return
		}
	}
}
