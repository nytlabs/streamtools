package main

import (
    "log"
    "encoding/json"
)

// TODO: dump entire object when exact match is required. 
// for example:
// if mask is {"id":"hello","date":null} 
// and input JSON is {"id":"goodbye","date": 1367432309000}
// the returned object will be {"date":1367432309000}
// all other cases currently work.

// filter should probably return a map[string]interface{} with a len of 0 if any exact match 
// is missed like in the above example

// recursively removes keys for JSON object
func filter(mask map[string]interface{}, input map[string]interface{}) map[string]interface{} {
    t := make(map[string]interface{})

    // iterate only over mask keys
    for k, v := range mask {
        switch vv := input[k].(type) {
        case []interface{}: // arrays
            t[k] = make([]interface{},0)
            for _, u := range vv {
                t[k] = append(t[k].([]interface{}), filter( mask[k].(map[string]interface{}), u.(map[string]interface{}) ) )
            }
        case map[string]interface{}: // object
            t[k] = filter( mask[k].(map[string]interface{}), input[k].(map[string]interface{}) )
        default: // string, int, float
            _, ok := input[k]
            if ok && v == nil || ok && v == input[k] {
                t[k] = input[k]
            }
        }
    }

    return t
}


func main(){

    testJSON := make([][]byte, 4)
    testJSON[0] = []byte(`{"id": "quijibo","posts":[{"title":"this is a zzz","sub":{"ok":1}, "p":1.52311231,"count":5,"url":"http://www.asdfasdfasdf.com"},{"title":"this is a zob","p":3.12311231,"count":5,"url":"http://www.asdfasdfasdf.com"}], "date":1367432309000, "obj":{"test":"ok"}}`)
    testJSON[1] = []byte(`{"id": "zomo","posts":[{"title":"this is a aaa","sub":{"ok":10}, "p":2.32111231,"count":6,"url":"http://www.111111111.com"},{"title":"this is a cog","p":5.552311231,"count":500,"url":"http://www.asdfasdfasdf.com"}], "date":1375381109000}`)
    testJSON[2] = []byte(`{"id": "slip","posts":[{"title":"this is a bbb","sub":{"ok":1000}, "p":-10.211231,"count":7,"url":"http://www.22222222.com"},{"title":"this is a zog","p":10.52311231,"count":50000,"url":"http://www.asdfasdfasdf.com"}], "date": 1375726709000}`)
    testJSON[3] = []byte(`{"garbage":"N/A"}`)
    //testJSON[4] = []byte(`totally broken`)

    testMask := make([][]byte, 7)
    testMask[0] = []byte(`{"id": null}`)
    testMask[1] = []byte(`{"id": "quijibo"}`)
    testMask[2] = []byte(`{"id": "quijibo", "date": null}`)
    testMask[3] = []byte(`{"id": null, "posts":{"url":null}}`)
    testMask[4] = []byte(`{"id": null, "posts":{"p":null}}`)
    testMask[5] = []byte(`{"posts":{"sub":{"ok":null}}}`)
    testMask[6] = []byte(`{"posts":{"sub":{"ok":1000}}}`)

    for _, tm := range testMask {
        log.Println(string(tm))
        for _, tj := range testJSON {
            var tjj interface{}
            var tmj interface{}
            err := json.Unmarshal(tj, &tjj)
            if err != nil{
                panic("BAD INPUT")
            }
            err = json.Unmarshal(tm, &tmj)
            if err != nil{
                panic("BAD PATTERN")
            }
            f := filter(tmj.(map[string]interface{}), tjj.(map[string]interface{}))
            log.Println(f)
        }
    }

}