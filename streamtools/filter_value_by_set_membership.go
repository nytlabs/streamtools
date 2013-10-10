package streamtools

import (
    "github.com/bitly/go-simplejson"
    "strings"
    "log"
)

func FilterValueBySetMembership(inChan chan *simplejson.Json, outChan chan *simplejson.Json, RuleChan chan *simplejson.Json) {
    set := make(map[string]interface{})
    rules := <-RuleChan

    fullPath, err := rules.Get("path").String()
    if err != nil{
        log.Fatalf("invalid key path")
    }
    pathS := strings.Split(fullPath, ".")

    setStr, err := rules.Get("set").String()
    if err != nil{
        log.Fatalf("invalid set")
    }

    setJson, err := simplejson.NewJson([]byte(setStr))
    if err != nil{
        log.Fatalf("could not make json out of set")
    }

    setS, err := setJson.StringArray()
    if err != nil{
        log.Fatalf("invalid set")
    }

    for _, t := range setS {
        set[t] = nil
    }

    for {
        select {
        case rules := <-RuleChan:
        case msg := <-inChan:
            val, err := msg.GetPath(pathS...).String()
            if err != nil{
                log.Println(msg)
                log.Println("could not find key")
            }

            if _, ok := set[val]; ok {
                outChan <- msg
            } 

        }
    }
}