package blocks

import (
	"time"
)

func Date(b *Block) {

	type dateRule struct {
		FmtString string
		Period    int
	}

	var rule *dateRule
	var d time.Duration

	timer := time.NewTimer(time.Duration(1) * time.Second)

	for {
		select {
		case t := <-timer.C:
			if rule == nil {
				break
			}

			outMsg := make(map[string]interface{})
			Set(outMsg, "date", t.Format(rule.FmtString))
			broadcast(b.OutChans, outMsg)
			timer.Reset(d)
		case msg := <-b.Routes["get_rule"]:
			if rule == nil {
				marshal(msg, &dateRule{})
			} else {
				marshal(msg, rule)
			}
		case msg := <-b.Routes["set_rule"]:
			if rule == nil {
				rule = &dateRule{}
			}
			unmarshal(msg, rule)

			d = time.Duration(rule.Period) * time.Second
			timer.Reset(d)

		case msg := <-b.AddChan:
			updateOutChans(msg, b)
		case <-b.QuitChan:
			quit(b)
			return
		}
	}
}
