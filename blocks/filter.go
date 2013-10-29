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

	rule := &filterRule{}
	unmarshal(<-b.Routes["set_rule"], &rule)

	for {
		select {
		case msg := <-b.InChan:
			values := getKeyValues(msg, rule.Path)
			for _, value := range values {
				if operators[rule.Operator](value, rule.Comparator) == !rule.Invert {
					broadcast(b.OutChans, msg)
					break
				}
			}
		case msg := <-b.Routes["set_rule"]:
			unmarshal(msg, &rule)
		case msg := <-b.Routes["get_rule"]:
			marshal(msg, rule)
		case msg := <-b.AddChan:
			updateOutChans(msg, b)
		case <-b.QuitChan:
			quit(b)
			return
		}
	}

}
