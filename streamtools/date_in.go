package streamtools

import (
	"github.com/bitly/go-simplejson"
	"log"
	"time"
)

func Date(outChan chan *simplejson.Json, RuleChan chan *simplejson.Json) {

	rule := <-RuleChan
	fmtString, err := rule.Get("fmtString").String()
	if err != nil {
		log.Fatal(err.Error())
	}

	now := time.Now()
	tomorrow := now.Add(time.Duration(12) * time.Hour).Round(time.Duration(24) * time.Hour)
	d := tomorrow.Sub(now)

	timer := time.NewTimer(d)

	msgJson, err := simplejson.NewJson([]byte("{}"))
	if err != nil {
		log.Fatal(err.Error())
	}

	for {
		select {
		case <-RuleChan:
		case t := <-timer.C:

			msgJson.Set("date", t.Format(fmtString))

		}
	}

}
