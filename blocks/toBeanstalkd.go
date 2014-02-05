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
    var conn *lentil.Beanstalkd
    var e error
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

            conn, e = lentil.Dial(rule.Host)
            if e != nil {
                log.Println(e.Error())
            } else {
                if len(rule.Tube) > 0 {
                    tube = rule.Tube
                }
                if rule.TTR > 0 {
                    ttr = rule.TTR
                }
                conn.Use(tube)
                log.Println("initialized connection using tube:", tube)
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
                break
            }
            /* your code goes here */
 			if conn == nil {
 				log.Panic("Connection to beanstalkd was dropped. Something is not right.")
 				break
 			}
            jobId, err := conn.Put(0, 0, ttr, msgStr)
            if err != nil {
                log.Println(err.Error())
            } else{
                log.Println("put job on queue: Job Id:", jobId)
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
