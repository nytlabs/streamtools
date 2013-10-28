package blocks

import (
	"github.com/bitly/go-simplejson"
	"log"
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
	log.Println(rule)

	timer := time.NewTimer(time.Duration(1) * time.Second)
	outMsg := simplejson.New()
	d := time.Duration(rule.Period) * time.Second

	for {
		select {
		case t := <-timer.C:
			outMsg.Set("date", t.Format(rule.FmtString))
			broadcast(b.OutChans, outMsg)
			timer.Reset(d)
		case msg := <-b.AddChan:
			updateOutChans(msg, b)
		case <-b.QuitChan:
			quit(b)
			return
		}
	}
}
