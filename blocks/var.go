package blocks

// Var calculates variance in an online fashion using Welford's Algorithm.
// The Var() block is the Sd() block with the squared correction.
// Ref: http://citeseerx.ist.psu.edu/viewdoc/download?doi=10.1.1.302.7503&rep=rep1&type=pdf
func Var(b *Block) {

	type varRule struct {
		Key string
	}

	type varData struct {
		Variance float64
	}

	data := &varData{Variance: 0.0}
	var rule *varRule

	N := 0.0
	M_curr := 0.0
	M_new := 0.0
	S := 0.0

	for {
		select {
		case query := <-b.Routes["var"]:
			marshal(query, data)
		case ruleUpdate := <-b.Routes["set_rule"]:
			if rule == nil {
				rule = &varRule{}
			}
			unmarshal(ruleUpdate, rule)
		case msg := <-b.Routes["get_rule"]:
			if rule == nil {
				marshal(msg, &varRule{})
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
			x, ok := val.(float64)
			if !ok {
				break
			}
			N++
			if N == 1.0 {
				M_curr = x
			} else {
				M_new = M_curr + (x-M_curr)/N
				S = S + (x-M_curr)*(x-M_new)
				M_curr = M_new
			}
			data.Variance = S / (N - 1.0)
		}
	}
}
