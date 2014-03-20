package library

import (
    "github.com/nytlabs/streamtools/st/blocks" // blocks
    "github.com/gorilla/websocket"
    "encoding/json"
    "time"
    "net/http"
)

// specify those channels we're going to use to communicate with streamtools
type FromWebsocket struct {
    blocks.Block
    queryrule chan chan interface{}
    inrule    chan interface{}
    inpoll    chan interface{}
    in        chan interface{}
    out       chan interface{}
    quit      chan interface{}
}

// we need to build a simple factory so that streamtools can make new blocks of this kind
func NewFromWebsocket() blocks.BlockInterface {
    return &FromWebsocket{}
}

// Setup is called once before running the block. We build up the channels and specify what kind of block this is.
func (b *FromWebsocket) Setup() {
    b.Kind = "FromWebsocket"
    b.inrule = b.InRoute("rule")
    b.queryrule = b.QueryRoute("rule")
    b.quit = b.Quit()
    b.out = b.Broadcast()
}

func recv(ws *websocket.Conn, out chan interface{}){
    for {
        _, p, err := ws.ReadMessage()
        if err != nil {
           return
        }

        var outMsg interface{}
        err = json.Unmarshal(p, &outMsg)
        if err != nil {
            continue
        }
        out <- outMsg
    }
}

// Run is the block's main loop. Here we listen on the different channels we set up.
func (b *FromWebsocket) Run() {
    var ws *websocket.Conn
    var URL string
    var handshakeDialer = &websocket.Dialer{
        Subprotocols:    []string{"p1", "p2"},
    }
    listenWS := make(chan interface{})
    wsHeader := http.Header{"Origin": {"http://localhost/"}}

    for {
        select {
        case ruleI := <-b.inrule:
            var err error
            // set a parameter of the block
            r, ok := ruleI.(map[string]interface{})
            if !ok {
                b.Error("bad rule")
                break
            }

            url, ok := r["url"]
            if !ok {
                b.Error("no url specified")
                break
            }
            surl, ok := url.(string)
            if !ok {
                b.Error("error reading url")
                break
            }
            if ws != nil {
                ws.Close()
            }

            ws, _, err = handshakeDialer.Dial(surl, wsHeader)          
            if err != nil {
                b.Error("could not connect to url")
                break
            }
            ws.SetReadDeadline(time.Time{})  
            go recv(ws, listenWS)

            URL = surl
        case <-b.quit:
            // quit the block
            return
        case o := <-b.queryrule:
            o <- map[string]interface{}{
                "url": URL,
            }
        case in := <- listenWS:
            b.out <- in
        }
    }
}
