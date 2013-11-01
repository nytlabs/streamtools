// Streaming Standard Deviation
// Welford's Algorithm

package blocks

import (
    "log"
)

func Sd(b *Block) {

    type sdRule struct {
        Key string
    }

    type sdData struct {
        Sd float64
    }

    rule := &sdRule{}
    data := &sdData{Sd: 0}

    // block unitl a rule is set
    unmarshal(<-b.Routes["set_rule"], &rule)

    data.Sd = 0.0
    N := 0.0

    for {
        select {
        case query := <-b.Routes["avg"]:
            marshal(query, data)
        case msg := <-b.Inchan:
            val := getKeyValues(msg, rule.Key)[0].(json.Number)
            x, err := val.Float64()
            if err != nil {
                log.Fatal(err.Error())
            }
            N = N + 1.0
            if N == 1.0 {
                M := x
                S := 0
            } else {
                M_new = M + (x - M) / N
                S = S + (x - M)*(x - M_new)
                M = M_new
            }
            data.Sd = S / (N - 1)
}
