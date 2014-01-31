package blocks

import (
	"time"
)

// emits the time. Specify the period - the time between emissions - in seconds
// as a rule.
func Ticker(b *Block) {

	type tickerRule struct {
		Interval string
	}

	rule := &tickerRule{
		Interval: "1s",
	}

	tickC := time.Tick(time.Duration(1) * time.Second)

	for {
		select {
		case tick := <-tickC:
			msg := make(map[string]interface{})
			Set(msg, "t", tick)
			out := BMsg{
				Msg:          msg,
				ResponseChan: nil,
			}
			broadcast(b.OutChans, &out)
		case msg := <-b.AddChan:
			updateOutChans(msg, b)

		case r := <-b.Routes["set_rule"]:
			unmarshal(r, rule)
			newDur, err := time.ParseDuration(rule.Interval)
			if err != nil {
				break
			}
			tickC = time.Tick(newDur)

		case r := <-b.Routes["get_rule"]:
			marshal(r, rule)

		case <-b.QuitChan:
			quit(b)
			return
		}
	}
}
