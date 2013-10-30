package blocks

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

			unmarshal(msg, rule)
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
