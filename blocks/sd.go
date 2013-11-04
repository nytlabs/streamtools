// Streaming Standard Deviation
// Welford's Algorithm
// http://amstat.tandfonline.com/doi/abs/10.1080/00401706.1962.10490022?journalCode=utch20#preview

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
        Sd float64
    }

    rule := &sdRule{}
    data := &sdData{Sd: 0}

    // block until a rule is set
    unmarshal(<-b.Routes["set_rule"], &rule)

    data.Sd = 0.0
    N := 0.0
    M_curr := 0.0
    M_new := 0.0
    S := 0.0

    for {
        select {
        case query := <-b.Routes["sd"]:
            marshal(query, data)
        case msg := <-b.InChan:
            val := getKeyValues(msg, rule.Key)[0].(json.Number)
            x, err := val.Float64()
            if err != nil {
                log.Fatal(err.Error())
            }
            N = N + 1.0
            if N == 1.0 {
                M_curr = x
            } else {
                M_new = M_curr + (x - M_curr) / N
                S = S + (x - M_curr)*(x - M_new)
                M_curr = M_new
            }
            data.Sd = math.Sqrt(S / (N - 1.0))
        }
    }
}
