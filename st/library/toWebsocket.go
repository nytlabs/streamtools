package library

import (
	"github.com/nytlabs/streamtools/st/blocks" // blocks
	"net/http"
)

// specify those channels we're going to use to communicate with streamtools
type towebsocket struct {
	blocks.Block
	queryrule chan chan interface{}
	inrule    chan interface{}
	inpoll    chan interface{}
	in        chan interface{}
	out       chan interface{}
	quit      chan interface{}
}

// we need to build a simple factory so that streamtools can make new blocks of this kind
func Newtowebsocket() blocks.BlockInterface {
	return &towebsocket{}
}

// Setup is called once before running the block. We build up the channels and specify what kind of block this is.
func (b *towebsocket) Setup() {
	b.Kind = "towebsocket"
	b.in = b.InRoute("in")
	b.inrule = b.InRoute("rule")
	b.queryrule = b.QueryRoute("rule")
	b.inpoll = b.InRoute("poll")
	b.quit = b.Quit()
	b.out = b.Broadcast()
}

// Run is the block's main loop. Here we listen on the different channels we set up.
func (b *towebsocket) Run() {

	var addr string

	go h.run()

	for {
		select {
		case ruleI := <-b.inrule:
			// set a parameter of the block
			rule, ok := ruleI.(map[string]interface{})
			addr := utils.ParseString(rule, "port")
			http.HandleFunc("/ws", serveWs)
			err := http.ListenAndServe(*addr, nil)
			if err != nil {
				b.Error(err)
			}
		case <-b.quit:
			// quit the block
			return
		case _ = <-b.in:
			// deal with inbound data
		case <-b.inpoll:
			// deal with a poll request
		case _ = <-b.queryrule:
			// deal with a query request
		}
	}
}

type hub struct {
	// Registered connections.
	connections map[*connection]bool

	// Inbound messages from the connections.
	broadcast chan []byte

	// Register requests from the connections.
	register chan *connection

	// Unregister requests from connections.
	unregister chan *connection
}

var h = hub{
	broadcast:   make(chan []byte),
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
					close(c.send)
					delete(h.connections, c)
				}
			}
		}
	}
}
