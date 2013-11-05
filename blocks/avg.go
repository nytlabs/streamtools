// Online average
// Update a sample average on a per point basis
// mu_i+1 = mu_i * (n - 1) /n + (1/n) * x_i

package blocks

import (
    "encoding/json"
    "log"
)

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
            val := getKeyValues(msg, rule.Key)[0].(json.Number)
            x, err := val.Float64()
            if err != nil {
                log.Fatal(err.Error())
            }
            N++
            data.Avg = ((N - 1.0) / N)*data.Avg + (1.0/N)*x
        }
    }
}
