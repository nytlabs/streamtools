package blocks

// rule - input
// data - output

func Avg(b *Block) {

    type avgRule struct {
        Key string
        X float64
    }

    type avgData {
        Avg float64
        N float64
    }

    rule := &avgRule{}
    data := &avgData{}

    // block until a rule is set
    unmarshal(<-b.Routes["set_rule"], &rule)

    data.Avg = 0.0
    data.N = 0.0

    for {
        select {
        case query := <-b.Routes["avg"]:
            marshal(query, data)
        case msg := <-b.InChan:
            val := getKeyValues(msg, rule.Key)[0].(float64)
            data.N = data.N + 1.0
            data.Avg = ((n - 1.0) / n)*data.Avg + (1/n)*val
        }
    }


}
