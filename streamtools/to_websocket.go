// https://gist.github.com/garyburd/1316852
package streamtools

import (
	"code.google.com/p/go.net/websocket"
	"github.com/bitly/go-simplejson"
	"log"
	"net/http"
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

// ToWebSocket emits all messages from a topic into a base64 encoded websocket handle
func ToWebSocket(inChan chan *simplejson.Json, RuleChan chan *simplejson.Json) {
	rules := <-RuleChan
	wsPort, err := rules.Get("wsport").String()
	wsHandle, err := rules.Get("wshandle").String()

	var wsConfig *websocket.Config
	var h = wsServer{
		broadcast:   make(chan string),
		register:    make(chan *wsConnection),
		unregister:  make(chan *wsConnection),
		connections: make(map[*wsConnection]bool),
	}

	wsConfig, err = websocket.NewConfig("ws://localhost:"+wsPort+"/", "http://localhost:"+wsPort+"/")
	if err != nil {
		log.Fatalf(err.Error())
	}

	wsConfig.Protocol = []string{"base64"}

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

	for {
		select {
		case <-RuleChan:
		case msg := <-inChan:
			m, err := msg.Encode()
			if err != nil {
				log.Println("could not encode JSON")
			}
			h.broadcast <- string(m)
		}
	}
}
