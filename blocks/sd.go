// Streaming Standard Deviation
// Welford's Algorithm
// http://citeseerx.ist.psu.edu/viewdoc/download?doi=10.1.1.302.7503&rep=rep1&type=pdf

package blocks

import (
    "encoding/json"
    "log"
    "math"
)

func Sd(b *Block) {

    type sdRule struct {
        Key string
    }

    type sdData struct {
        StDev float64
    }

    data := &sdData{StDev: 0.0}
    var rule *sdRule

    N := 0.0
    M_curr := 0.0
    M_new := 0.0
    S := 0.0

    for {
        select {
        case query := <-b.Routes["sd"]:
            marshal(query, data)
        case ruleUpdate := <-b.Routes["set_rule"]:
            if rule == nil {
                rule = &sdRule{}
            }
            unmarshal(ruleUpdate, rule)
        case msg := <-b.Routes["get_rule"]:
            if rule == nil {
                marshal(msg, &sdRule{})
            } else {
                marshal(msg, rule)
            }
        case <-b.QuitChan:
            quit(b)
            return
        case msg := <-b.InChan:
            val := getKeyValues(msg, rule.Key)[0].(json.Number)
            x, err := val.Float64()
            if err != nil {
                log.Fatal(err.Error())
            }
            N++
            if N == 1.0 {
                M_curr = x
            } else {
                M_new = M_curr + (x - M_curr) / N
                S = S + (x - M_curr)*(x - M_new)
                M_curr = M_new
            }
            data.StDev = math.Sqrt(S / (N - 1.0))
        }
    }
}
