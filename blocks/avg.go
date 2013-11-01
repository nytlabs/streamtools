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

    rule := &avgRule{}
    data := &avgData{Avg: 0}

    // block until a rule is set
    unmarshal(<-b.Routes["set_rule"], &rule)

    data.Avg = 0.0
    N := 0.0

    for {
        select {
        case query := <-b.Routes["avg"]:
            marshal(query, data)
        case msg := <-b.InChan:
            val := getKeyValues(msg, rule.Key)[0].(json.Number)
            val_f, err := val.Float64()
            if err != nil {
                log.Fatal(err.Error())
            }
            N = N + 1.0
            data.Avg = ((N - 1.0) / N)*data.Avg + (1.0/N)*val_f
        }
    }
}
