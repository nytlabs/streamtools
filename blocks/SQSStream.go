package blocks

import (
        "github.com/mikedewar/aws4"
        "log"
        "time"
        "net/url"
        "io/ioutil"
        "encoding/xml"
        "encoding/json"
)

var (
        AWSSQSAPIVersion    string = "2012-11-05"
        AWSSignatureVersion string = "4"
)

type fromSQSRule struct {
        SQSEndpoint   string
        AccessKey     string
        AccessSecret  string
}


type Message struct {
        // this is a list in case I'm ever brave enough to up the "MaxNumberOfMessages" away from 1
        Body          []string `xml:"ReceiveMessageResult>Message>Body"`
        ReceiptHandle []string `xml:"ReceiveMessageResult>Message>ReceiptHandle"`
}

func pollSQS(rule *fromSQSRule) Message {
        query := make(url.Values)
        query.Add("Action", "ReceiveMessage")
        query.Add("AttributeName", "All")
        query.Add("Version", AWSSQSAPIVersion)
        query.Add("SignatureVersion", AWSSignatureVersion)
        query.Add("WaitTimeSeconds", "10")

        keys := &aws4.Keys{
                AccessKey: rule.AccessKey,
                SecretKey: rule.AccessSecret,
        }

        c := aws4.Client{Keys: keys}

        resp, err := c.Get(rule.SQSEndpoint + query.Encode())
        if err != nil {
                log.Println(err.Error())
        }

        var v Message

        body, err := ioutil.ReadAll(resp.Body)
        err = xml.Unmarshal(body, &v)
        if err != nil {
                log.Println(err.Error())
        }
        return v
}

func deleteMessage(rule *fromSQSRule, ReceiptHandle string) {
        query := make(url.Values)
        query.Add("Action", "DeleteMessage")
        query.Add("ReceiptHandle", ReceiptHandle)
        query.Add("Version", AWSSQSAPIVersion)
        query.Add("SignatureVersion", AWSSignatureVersion)

        keys := &aws4.Keys{
                AccessKey: rule.AccessKey,
                SecretKey: rule.AccessSecret,
        }

        c := aws4.Client{Keys: keys}

        _, err := c.Get(rule.SQSEndpoint + query.Encode())
        if err != nil {
                log.Println(err.Error())
        }
}

func FromSQS(b *Block) {
        var rule *fromSQSRule
        timer := time.NewTimer(1)
        
        for {
                select {
                case <-timer.C:
                        if rule == nil {
                                timer.Reset(time.Duration(10) * time.Second)
                                break
                        }

                        m := pollSQS(rule)
                        if len(m.Body) > 0 {
                                for i, body := range m.Body{
                                        var msg BMsg
                                        err := json.Unmarshal([]byte(body), &msg)
                                        if err != nil {
                                                log.Println(err.Error())
                                        }
                                        broadcast(b.OutChans, msg)
                                        deleteMessage(rule, m.ReceiptHandle[i])
                                }
                                timer.Reset(time.Duration(10) * time.Millisecond)
                        } else {
                                timer.Reset(time.Duration(10) * time.Second)
                        }

                case msg := <-b.Routes["set_rule"]:
                        unmarshal(msg, rule)
                case msg := <-b.Routes["get_rule"]:
                        if rule == nil {
                                marshal(msg, &fromSQSRule{})
                        } else {
                                marshal(msg, rule)
                        }
                case msg := <-b.AddChan:
                        updateOutChans(msg, b)
                case <-b.QuitChan:
                        quit(b)
                        return
                }
        }
}