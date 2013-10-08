package streamtools

import (
	"github.com/bitly/go-simplejson"
	"log"
	"time"
)

func Date(outChan chan *simplejson.Json, RuleChan chan *simplejson.Json) {
	timer := time.NewTimer(time.Duration(0) * time.Second)
	utcLoc, _ := time.LoadLocation("UTC")
	emit := make(chan time.Time, 1)
	msgJson, _ := simplejson.NewJson([]byte("{}"))
	rule := <-RuleChan
	fmtString, err := rule.Get("fmtString").String()
	if err != nil {
		log.Fatal(err.Error())
	}

	for {
		select {
		case <-RuleChan:
		case <-timer.C:
			emit <- time.Now()
		case t := <-emit:
			msgJson.Set("date", t.Format(fmtString))
			log.Println("[DATEIN] emitting", msgJson)
			outChan <- msgJson

			tomorrow := time.Date(t.Year(), t.Month(), t.Day()+1, 0, 0, 0, 0, utcLoc)
			d := tomorrow.Sub(t)
			timer.Reset(d)
			log.Println("[DATEIN] Next emit time in", d)
		}
	}
}
