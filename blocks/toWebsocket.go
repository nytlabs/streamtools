package blocks

import (
	"code.google.com/p/go.net/websocket"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
)

type wsConnection struct {
	ws   *websocket.Conn
	send chan string
}

func (c *wsConnection) reader(h *wsServer) {
	for {
		var message string
		err := websocket.Message.Receive(c.ws, &message)
		if err != nil {
			break
		}
		h.broadcast <- message
	}
	c.ws.Close()
}

func (c *wsConnection) writer(h *wsServer) {
	for message := range c.send {
		err := websocket.Message.Send(c.ws, message)
		if err != nil {
			break
		}
	}
	c.ws.Close()
}

type wsServer struct {
	connections map[*wsConnection]bool
	broadcast   chan string
	register    chan *wsConnection
	unregister  chan *wsConnection
}

func (h *wsServer) run() {
	for {
		select {
		case c := <-h.register:
			h.connections[c] = true
		case c := <-h.unregister:
			delete(h.connections, c)
			close(c.send)
		case m := <-h.broadcast:
			for c := range h.connections {
				select {
				case c.send <- m:
				default:
					delete(h.connections, c)
					close(c.send)
					go c.ws.Close()
				}
			}
		}
	}
}

// ToWebsocket stands up and writes to a websocket
func ToWebsocket(b *Block) {

	var err error

	type wsRule struct {
		Port     int
		Endpoint string
	}

	var rule *wsRule

	var h = wsServer{
		broadcast:   make(chan string),
		register:    make(chan *wsConnection),
		unregister:  make(chan *wsConnection),
		connections: make(map[*wsConnection]bool),
	}

	var wsConfig *websocket.Config

	for {
		select {
		case msg := <-b.AddChan:
			updateOutChans(msg, b)
		case msg := <-b.Routes["set_rule"]:
			if rule == nil {
				rule = &wsRule{}
			}
			unmarshal(msg, rule)

			wsPort := strconv.Itoa(rule.Port)
			wsConfig, err = websocket.NewConfig("ws://localhost:"+wsPort+"/", "http://localhost:"+wsPort+"/")
			if err != nil {
				log.Fatalf(err.Error())
			}

			wsHandle := rule.Endpoint
			if !strings.HasPrefix(wsHandle, "/") {
				wsHandle = "/" + wsHandle
			}
			http.Handle(wsHandle, websocket.Server{
				Handler: func(ws *websocket.Conn) {
					c := &wsConnection{send: make(chan string), ws: ws}
					h.register <- c
					defer func() { h.unregister <- c }()
					go c.writer(&h)
					c.reader(&h)
				},
				Config: *wsConfig,
			})

			go h.run()

			go func() {
				if err := http.ListenAndServe("localhost:"+wsPort, nil); err != nil {
					log.Fatal("ListenAndServe:", err)
				}
			}()

		case msg := <-b.Routes["get_rule"]:
			if rule == nil {
				marshal(msg, &wsRule{})
			} else {
				marshal(msg, rule)
			}
		case <-b.QuitChan:
			quit(b)
			return
		case msg := <-b.InChan:
			if rule == nil {
				break
			}

			postBody, err := json.Marshal(msg.Msg)
			if err != nil {
				log.Fatal(err.Error())
				break
			}

			h.broadcast <- string(postBody)
		}
	}
}
