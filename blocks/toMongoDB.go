package blocks

import (
    "log"
    "encoding/json"
    "labix.org/v2/mgo"
    //"labix.org/v2/mgo/bson"
)

func ToMongoDB(b *Block) {

    type toMongoDBRule struct {
        Host string
        Database string
        Collection string
    }


    var rule *toMongoDBRule
    var collection *mgo.Collection
    //var session  * mgo.Session
    //var db *mgo.Database
    //var db = "test"


    for {
        select {
        case m := <-b.Routes["set_rule"]:
            if rule == nil {
                rule = &toMongoDBRule{}
            }
            unmarshal(m, rule)

            session, err := mgo.Dial(rule.Host)
            if err != nil {
                log.Println("Could not connect to MongoDB", err.Error())
                break
            } 
            if len(rule.Database) <=0 {
                log.Println("Database field is empty")
                break
            } 
            if len(rule.Collection) <= 0 {
                log.Println("Collection name is empty")
                break
            }
            collection = session.DB(rule.Database).C(rule.Collection)

        case r := <-b.Routes["get_rule"]:
            if rule == nil {
                marshal(r, &toMongoDBRule{})
            } else {
                marshal(r, rule)
            }

        case msg := <-b.InChan:
            if rule == nil {
                break
            }
            msgStr, err := json.Marshal(msg.Msg)
            if err != nil {
                log.Println("wow bad json" , err.Error())
                break
            }
            var m map[string]interface{}
            err = json.Unmarshal(msgStr, &m)
            if err != nil {
                log.Println("wow bad json" , err.Error())
                break
            }
            err = collection.Insert(m)
            
            
        case msg := <-b.AddChan:
            updateOutChans(msg, b)
        case <-b.QuitChan:
            quit(b)
            return
        }
    }
}
