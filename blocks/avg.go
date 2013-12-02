package blocks

import "log"

// Avg() is an online average
// The average for a stream of data is updated 1 data point at a time.
// Formula: mu_i+1 = mu_i * (n - 1) /n + (1/n) * x_i
func Avg(b *Block) {

	type avgRule struct {
		Key string
	}

	type avgData struct {
		Avg float64
	}

	data := &avgData{Avg: 0.0}
	var rule *avgRule

	N := 0.0

	for {
		select {
		case query := <-b.Routes["avg"]:
			marshal(query, data)
		case ruleUpdate := <-b.Routes["set_rule"]:
			if rule == nil {
				rule = &avgRule{}
			}
			unmarshal(ruleUpdate, rule)
		case msg := <-b.Routes["get_rule"]:
			if rule == nil {
				marshal(msg, &avgRule{})
			} else {
				marshal(msg, rule)
			}
		case <-b.QuitChan:
			quit(b)
			return
		case msg := <-b.InChan:
			if rule == nil {
				break
			}
			val := getKeyValues(msg, rule.Key)[0]
			var x float64
			switch val := val.(type) {
			case float64:
				x = val
			case int:
				x = float64(val)
			default:
				log.Println("unable to take average of value specified by", rule.Key)
				break
			}
			N++
			data.Avg = ((N-1.0)/N)*data.Avg + (1.0/N)*x
		}
	}
}
