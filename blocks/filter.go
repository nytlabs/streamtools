package blocks

type opFunc func(interface{}, interface{}) bool

var operators map[string]opFunc

func Filter(b *Block) {

	type filterRule struct {
		Operator   string
		Path       string
		Comparator interface{}
	}

	operators = make(map[string]opFunc)

	operators["="] = equals
	/*
		operators[">"] = greaterthan
		operators["<"] = lessthan
		operators["∈"] = elementof
		operators["⊂"] = subsetof
	*/

	rule := &filterRule{}
	unmarshal(<-b.Routes["set_rule"], &rule)

	for {
		select {
		case msg := <-b.InChan:
			value := getPath(msg, rule.Path)
			if operators[rule.Operator](value, rule.Comparator) {
				broadcast(b.OutChans, msg)
			}
		case msg := <-b.AddChan:
			updateOutChans(msg, b)
		}
	}

}
