package blocks

import (
	"time"
)

func Ticker(b *Block) {

	type tickerRule struct {
		Period int
	}

	rule := &tickerRule{
		Period: 1,
	}

	ticker := time.NewTicker(time.Duration(rule.Period) * time.Second)

	for {
		select {
		case tick := <-ticker.C:
			var msg BMsg
			Set(msg, "t", tick)
			broadcast(b.OutChans, msg)

		case msg := <-b.AddChan:
			updateOutChans(msg, b)

		case r := <-b.Routes["set_rule"]:
			unmarshal(r, &rule)
			ticker = time.NewTicker(time.Duration(rule.Period) * time.Second)

		case r := <-b.Routes["get_rule"]:
			marshal(r, rule)

		case <-b.QuitChan:
			quit(b)
			return
		}
	}
}
