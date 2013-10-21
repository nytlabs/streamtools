package blocks

import (
	"github.com/bitly/go-simplejson"
	"time"
)

func Date(b *Block) {

	type dateRule struct {
		FmtString string
		Period    int
	}

	rule := &dateRule{}

	// block until we recieve a rule
	unmarshal(<-b.Routes["set_rule"], &rule)

	timer := time.NewTimer(time.Duration(1) * time.Second)
	outMsg := simplejson.New()
	d := time.Duration(rule.Period) * time.Second

	for {
		select {
		case t := <-timer.C:
			outMsg.Set("date", t.Format(rule.FmtString))
			broadcast(b.OutChans, outMsg)
			timer.Reset(d)
		}
	}
}
