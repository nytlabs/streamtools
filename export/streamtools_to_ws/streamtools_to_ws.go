package main

import (
	"code.google.com/p/go.net/websocket"
	"flag"
	"github.com/bitly/nsq/nsq"
	"log"
	"net/http"
)

var (
	inTopic          = flag.String("in_topic", "", "topic to read from")
	channelName      = "ws_" + *inTopic
	lookupdHTTPAddrs = "127.0.0.1:4161"
	httpAddress      = flag.String("http-address", "0.0.0.0:8080", "<addr>:<port> to listen on for HTTP clients")
	maxInFlight      = flag.Int("max-in-flight", 100, "max number of messages to allow in flight")
)

type SyncHandler struct{}

func (self *SyncHandler) HandleMessage(m *nsq.Message) error {
	h.broadcast <- string(m.Body)
	return nil
}

// https://gist.github.com/garyburd/1316852
type connection struct {
	ws   *websocket.Conn
	send chan string
}

func (c *connection) reader() {
	for {
		var message string
		err := websocket.Message.Receive(c.ws, &message)
		if err != nil {
			break
		}
	}
	c.ws.Close()
}

func (c *connection) writer() {
	for message := range c.send {
		err := websocket.Message.Send(c.ws, message)
		if err != nil {
			break
		}
	}
	c.ws.Close()
}

func wsHandler(ws *websocket.Conn) {
	c := &connection{send: make(chan string, 256), ws: ws}
	h.register <- c
	defer func() { h.unregister <- c }()
	go c.writer()
	c.reader()
}

type hub struct {
	connections map[*connection]bool
	broadcast   chan string
	register    chan *connection
	unregister  chan *connection
}

var h = hub{
	broadcast:   make(chan string),
	register:    make(chan *connection),
	unregister:  make(chan *connection),
	connections: make(map[*connection]bool),
}

func (h *hub) run() {
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

func main() {
	flag.Parse()

	go h.run()

	r, err := nsq.NewReader(*inTopic, channelName)
	if err != nil {
		log.Fatal("could not make reader")
	}

	r.SetMaxInFlight(*maxInFlight)
	r.AddHandler(&SyncHandler{})

	err = r.ConnectToLookupd(lookupdHTTPAddrs)
	if err != nil {
		log.Fatal("could not connect to nsq")
	}

	http.Handle("/ws", websocket.Handler(wsHandler))

	if err = http.ListenAndServe(*httpAddress, nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}

	<-r.ExitChan
}
