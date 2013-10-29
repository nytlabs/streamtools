package blocks

func Timeseries(b *Block) {

	type tsRule struct {
		NumSamples int
		Key        string
	}

	type tsData struct {
		Values []float64
	}

	rule := &tsRule{}
	data := &tsData{}

	// block until we recieve a rule
	unmarshal(<-b.Routes["set_rule"], &rule)

	data.Values = make([]float64, rule.NumSamples)

	for {
		select {
		case query := <-b.Routes["timeseries"]:
			marshal(query, data)
		case ruleUpdate := <-b.Routes["set_rule"]:
			unmarshal(ruleUpdate, &rule)
		case msg := <-b.InChan:
			val := getKeyValues(msg, rule.Key)[0].(float64)
			data.Values = append(data.Values[1:], val)
		}
	}
}
