package streamtools

import (
    "github.com/bitly/go-simplejson"
)

func LastState(inChan chan *simplejson.Json, ruleChan chan *simplejson.Json, queryChan chan stateQuery) {
    lastMsg, _ := simplejson.NewJson([]byte(`{}`))
    for {
        select {
        case params = <-ruleChan:
        case query = <-queryChan:
            query.responseChan <- lastMsg
        case lastMsg = <-inChan:
        }
    }
}
