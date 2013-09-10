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
	msgJson, err := simplejson.NewJson([]byte("{}"))

	now := time.Now()
	// emit today immediately TODO is this cool?
	today := now.Add(-time.Duration(12) * time.Hour).Round(time.Duration(24) * time.Hour)
	msgJson.Set("date", today.Format(fmtString))
	log.Println("emitting", msgJson)
	outChan <- msgJson
	tomorrow := now.Add(time.Duration(12) * time.Hour).Round(time.Duration(24) * time.Hour)
	d := tomorrow.Sub(now)

	timer := time.NewTimer(d)

	if err != nil {
		log.Fatal(err.Error())
	}

	for {
		select {
		case <-RuleChan:
		case t := <-timer.C:

			msgJson.Set("date", t.Format(fmtString))
			outChan <- msgJson

		}
	}

}
