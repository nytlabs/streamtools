package streamtools

import (
	"github.com/bitly/go-simplejson"
	"log"
	"time"
)

type summary struct {
	count    int
	lastSeen time.Time
}

func MonitorKey(inChan chan *simplejson.Json, ruleChan chan *simplejson.Json, queryChan chan stateQuery) {

	rules := <-ruleChan
	key, err := rules.Get("key").String()
	if err != nil {
		log.Fatal(err)
	}

	set := make(map[string]*summary)

	for {
		select {
		case msg := <-inChan:

			k, err := getKey(key, msg).String()
			if err != nil {
				log.Fatal(err.Error())
			}
			if val, ok := set[k]; ok {
				val.count += 1
				val.lastSeen = time.Now()
			} else {
				val := summary{
					count:    1,
					lastSeen: time.Now(),
				}
				set[k] = &val
			}
		case query := <-queryChan:
			out, err := simplejson.NewJson([]byte("{}"))
			if err != nil {
				log.Fatal(err.Error())
			}
			out.Set("key", key)
			data := make([]*simplejson.Json, len(set))
			i := 0
			for k, v := range set {
				val, _ := simplejson.NewJson([]byte("{}"))
				val.Set("value", k)
				val.Set("count", v.count)
				val.Set("lastSeen", v.lastSeen)
				data[i] = val
				i += 1
			}
			out.Set("data", data)
			query.responseChan <- out
		}
	}
}
