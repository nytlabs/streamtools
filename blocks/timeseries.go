package blocks

func Timeseries(b *Block) {

	type tsRule struct {
		NumSamples int
		Key        string
	}

	type tsData struct {
		Values []float64
	}

	var rule *tsRule
	var data *tsData

	for {
		select {
		case query := <-b.Routes["timeseries"]:
			if data != nil {
				marshal(query, data)
			} else {
				marshal(query, &tsData{})
			}

		case ruleUpdate := <-b.Routes["set_rule"]:
			if rule == nil {
				rule = &tsRule{}
				data = &tsData{}
			}

			unmarshal(ruleUpdate, rule)
			data.Values = make([]float64, rule.NumSamples)

		case msg := <-b.Routes["get_rule"]:
			if rule == nil {
				marshal(msg, &tsRule{})
			} else {
				marshal(msg, rule)
			}
		case msg := <-b.InChan:
			if rule == nil {
				break
			}

			var val float64
			switch v := getKeyValues(msg, rule.Key)[0].(type) {
			case float32:
				val = float64(v)
			case int:
				val = float64(v)
			case float64:
				val = v
			}

			data.Values = append(data.Values[1:], val)
		}
	}
}
