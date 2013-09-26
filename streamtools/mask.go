package streamtools

import (
    "github.com/bitly/go-simplejson"
    "log"
)

func MaskJSON(mask map[string]interface{}, input map[string]interface{}) map[string]interface{} {
    t := make(map[string]interface{})
    for k, _ := range mask {
        switch input[k].(type) {
        case map[string]interface{}: 
            t[k] = MaskJSON(mask[k].(map[string]interface{}), input[k].(map[string]interface{}))
        default:
            t[k] = input[k]
        }
    }
    return t
}

func Mask(inChan chan *simplejson.Json, outChan chan *simplejson.Json, RuleChan chan *simplejson.Json) {
    mask := make(map[string]interface{})
    for {
        select {
        case inputRule := <-RuleChan:
            m, err := inputRule.Map()
            if err != nil{
                log.Fatalf(err.Error())
            }
            mask = m

        case msg := <-inChan:
            m, err := msg.Map()
            if err != nil{
                log.Fatalf(err.Error())
            }

            var t interface{}
            t = MaskJSON(mask, m)
            outChan <- t.(*simplejson.Json)
        }
    }
}