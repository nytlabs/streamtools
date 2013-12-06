package blocks

import "time"

func Timeseries(b *Block) {

	type tsRule struct {
		NumSamples int
		Key        string
		Lag        int
	}

	type tsDataPoint struct {
		Timestamp float64
		Value     float64
	}

	type tsData struct {
		Values []tsDataPoint
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
			data.Values = make([]tsDataPoint, rule.NumSamples)

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

      lag := -time.Duration(rule.Lag)*time.Second
      t := float64(time.Now().Add(lag).Unix())

			d := tsDataPoint{
				Timestamp: t,
				Value:     val,
			}
			data.Values = append(data.Values[1:], d)
		}
	}
}
