package library

import (
    "github.com/nytlabs/streamtools/st/blocks" // blocks
    "net/http"
    "encoding/json"
    "github.com/nytlabs/streamtools/st/util"
    "bytes"
    "io/ioutil"
)

// specify those channels we're going to use to communicate with streamtools
type PostHTTP struct {
    blocks.Block
    queryrule chan blocks.MsgChan
    inrule    blocks.MsgChan
    inpoll    blocks.MsgChan
    in        blocks.MsgChan
    out       blocks.MsgChan
    quit      blocks.MsgChan
}

// we need to build a simple factory so that streamtools can make new blocks of this kind
func NewPostHTTP() blocks.BlockInterface {
    return &PostHTTP{}
}

// Setup is called once before running the block. We build up the channels and specify what kind of block this is.
func (b *PostHTTP) Setup() {
    b.Kind = "PostHTTP"
    b.Desc = "POSTs messages to a given URL."
    b.in = b.InRoute("in")
    b.inrule = b.InRoute("rule")
    b.queryrule = b.QueryRoute("rule")
    b.out = b.Broadcast()
    b.quit = b.Quit()
}

// Run is the block's main loop. Here we listen on the different channels we set up.
func (b *PostHTTP) Run() {
    var url string
    contentType := "application/json"

    transport := http.Transport{
        Dial: dialTimeout,
    }

    client := &http.Client{
        Transport: &transport,
    }

    for {
        select {
        case ruleI := <-b.inrule:
            urlTmp, err := util.ParseString(ruleI, "Url")
            if err != nil {
                b.Error(err)
                break
            }

            contentTmp, err := util.ParseString(ruleI, "ContentType")
            if err != nil {
                b.Error(err)
                break
            }

            url = urlTmp
            contentType = contentTmp
        case <-b.quit:
            return
        case m := <-b.in:
            postBody, err := json.Marshal(m)
            if err != nil {
                b.Error(err)
                break
            }

            req, err := http.NewRequest("POST", url, bytes.NewReader(postBody))
            if err != nil {
                b.Error(err)
                break
            }

            resp, err := client.Do(req)
            if err != nil {
                b.Error(err)
                break
            }

            body, err := ioutil.ReadAll(resp.Body)

            b.out <- map[string]interface{}{
                "Response": string(body),
            }

            resp.Body.Close()
        case resp := <-b.queryrule:
            resp <- map[string]interface{}{
                "Url":url,
                "ContentType":contentType,
            }
        }
    }
}
