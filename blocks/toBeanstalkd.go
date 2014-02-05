package blocks

import (
    "log"
    "encoding/json"
    "github.com/nutrun/lentil"
)

func ToBeanstalkd(b *Block) {

    type toBeanstalkdRule struct {
        Host string
        Tube string
        TTR int
    }
    var conn lentil.Beanstalkd
    var tube = "default"
    var ttr = 0
    var rule *toBeanstalkdRule

    for {
        select {
        case m := <-b.Routes["set_rule"]:
            if rule == nil {
                rule = &toBeanstalkdRule{}
            }
            unmarshal(m, rule)
            conn, err := lentil.Dial(rule.Host)
            if err != nil {
                log.Panic(err.Error())
            } else {
                if len(rule.Tube) > 0 {
                    tube = rule.Tube
                }
                if rule.TTR > 0 {
                    ttr = rule.TTR
                }
                conn.Use(tube)
            }

        case r := <-b.Routes["get_rule"]:
            if rule == nil {
                marshal(r, &toBeanstalkdRule{})
            } else {
                marshal(r, rule)
            }

        case msg := <-b.InChan:
            if rule == nil {
                break
            }
            msgStr, err := json.Marshal(msg.Msg)
            if err != nil {
                log.Println("wow bad json")
            }
            /* your code goes here */
            jobId, err := conn.Put(0, 0, ttr, []byte(msgStr))
            if err != nil {
                log.Panic(err.Error())
            } else{
                log.Println("put jobId %d on queue", jobId)
            }
            //broadcast(b.OutChans, msg)
        case msg := <-b.AddChan:
            updateOutChans(msg, b)
        case <-b.QuitChan:
            quit(b)
            return
        }
    }
}
