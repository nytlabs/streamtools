package streamtools

import (
	"github.com/bitly/go-simplejson"
	"log"
	"math"
)

var (
	params simplejson.Json
	key    string
	query  stateQuery
	minval float64
)

func Min(inChan chan simplejson.Json, ruleChan chan simplejson.Json, queryChan chan stateQuery) {
	// initialise the min at +ve infinity
	minval = math.Inf(1)
	// block until we recieve a rule
	params = <-ruleChan
	key, err := params.Get("key").String()
	if err != nil {
		log.Fatal(err.Error())
	}
	for {
		select {
		case params = <-ruleChan:
			key, err = params.Get("key").String()
			if err != nil {
				log.Fatal(err.Error())
			}
		case query = <-queryChan:
			out, err := simplejson.NewJson([]byte{})
			if err != nil {
				log.Fatal(err.Error())
			}
			out.Set("min", minval)
		case msg := <-inChan:
			val, err := msg.Get(key).Float64()
			if err != nil {
				log.Fatal(err.Error())
			}
			minval = math.Min(val, minval)
		}
	}

}
