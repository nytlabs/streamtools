package daemon

import (
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"net/http"
	"log"
	"fmt"
	"encoding/json"
)

const (
	ADD_CHAN = iota
	DEL_CHAN 
)

var uiStream = hub{
	Broadcast:   make(chan []byte),
	register:    make(chan *connection),
	unregister:  make(chan *connection),
	connections: make(map[*connection]bool),
}

var logStream = hub{
	Broadcast:   make(chan []byte),
	register:    make(chan *connection),
	unregister:  make(chan *connection),
	connections: make(map[*connection]bool),
}

type Daemon struct {
	manager *BlockManager
	log chan *LogMsg
	ui  chan *LogMsg
	Port string
}

func NewDaemon() *Daemon {
	return &Daemon{
		manager: NewBlockManager(),
		log: make(chan *LogMsg),
		ui: make(chan *LogMsg),
	}
}

func addDefaultHeaders(fn http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Access-Control-Allow-Origin", "*")
        fn(w, r)
    }
}

func (d *Daemon) rootHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello!")
}

func (d *Daemon) staticHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello!")
}

func (d *Daemon) serveLogStream(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}
	w.Header().Set("Access-Control-Allow-Origin", "*")
	/*if r.Header.Get("Origin") != "http://"+r.Host {
		http.Error(w, "Origin not allowed", 403)
		return
	}*/
	ws, err := websocket.Upgrade(w, r, nil, 1024, 1024)
	if _, ok := err.(websocket.HandshakeError); ok {
		http.Error(w, "Not a websocket handshake", 400)
		return
	} else if err != nil {
		log.Println(err)
		return
	}
	c := &connection{send: make(chan []byte, 256), ws: ws, Hub: logStream}
	c.Hub.register <- c
	go c.writePump()
	c.readPump()
}

func (d *Daemon) serveUIStream(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}
	w.Header().Set("Access-Control-Allow-Origin", "*")
	/*if r.Header.Get("Origin") != "http://"+r.Host {
		http.Error(w, "Origin not allowed", 403)
		return
	}*/
	ws, err := websocket.Upgrade(w, r, nil, 1024, 1024)
	if _, ok := err.(websocket.HandshakeError); ok {
		http.Error(w, "Not a websocket handshake", 400)
		return
	} else if err != nil {
		log.Println(err)
		return
	}
	c := &connection{send: make(chan []byte, 256), ws: ws, Hub: uiStream}
	c.Hub.register <- c
	go c.writePump()
	c.readPump()
}

func (d *Daemon) importHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello!")
}

func (d *Daemon) exportHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello!")
}

func (d *Daemon) listBlockHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello!")
}

func (d *Daemon) createBlockHandler(w http.ResponseWriter, r *http.Request) {
	d.ui <- &LogMsg{
	    Type: "INFO",
	    Data: "{LASDLKASJDLKASJDASD}",
	    Id: "DAEMON",
	}

	d.log <- &LogMsg{
	    Type: "INFO",
	    Data: "BLOCK CREATED",
	    Id: "DAEMON",
	}
}

func (d *Daemon) blockInfoHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello!")
}

func (d *Daemon) deleteBlockHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello!")
}

func (d *Daemon) sendRouteHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello!")
}

func (d *Daemon) queryRouteHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello!")
}

func (d *Daemon) listConnectionHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello!")
}

func (d *Daemon) createConnectionHandler(w http.ResponseWriter, r *http.Request) {
	
}

func (d *Daemon) connectionInfoHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	conn, err := d.manager.GetConnection(vars["id"])
	if err != nil {
		ApiResponse(w, 500, "Connection not found.")
		return
	}

	jconn, err = json.Marshal(conn)
	if err != nil {
		ApiResponse(w, 500, "Error marshalling.")
		return
	}

	DataResponse(w, jconn)
}

func ApiResponse(w *http.ResponseWriter, statusCode int, statusTxt string) {
	response, err := json.Marshal(struct {
		StatusTxt string `json:"daemon"`
	}{
		statusTxt,
	})
	if err != nil {
		response = []byte(fmt.Sprintf(`{"daemon":"%s"}`, err.Error()))
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Content-Length", strconv.Itoa(len(response)))
	w.WriteHeader(statusCode)
	w.Write(response)
}

func DataResponse(w *rest.ResponseWriter, response []byte) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Content-Length", strconv.Itoa(len(response)))
	w.WriteHeader(200)
	w.Write(response)
}

func (d *Daemon)  deleteConnectionHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello!")
}


func (d *Daemon) Run(){
	go BroadcastStreams(d.ui, d.log)
	go uiStream.run()
	go logStream.run()

	r := mux.NewRouter()
	r.HandleFunc("/", d.rootHandler)
	r.HandleFunc("/static/{file}", d.staticHandler)
	r.HandleFunc("/log", d.serveLogStream)
	r.HandleFunc("/ui", d.serveUIStream)
	r.HandleFunc("/import", d.importHandler).Methods("POST")
	r.HandleFunc("/export", d.exportHandler).Methods("GET")
	r.HandleFunc("/blocks", d.listBlockHandler).Methods("GET") // list all blocks
	r.HandleFunc("/blocks", d.createBlockHandler).Methods("POST")
	r.HandleFunc("/blocks/{id}", d.createBlockHandler).Methods("POST") //create block
	r.HandleFunc("/blocks/{id}", d.blockInfoHandler).Methods("GET") // get block info
	r.HandleFunc("/blocks/{id}", d.deleteBlockHandler).Methods("DELETE") // delete block
	r.HandleFunc("/blocks/{id}/{route}", d.sendRouteHandler).Methods("POST") // send to block route
	r.HandleFunc("/blocks/{id}/{route}", d.queryRouteHandler).Methods("GET") // get from block route
	r.HandleFunc("/connections", d.createConnectionHandler).Methods("POST") // create connection
	r.HandleFunc("/connections", d.listConnectionHandler).Methods("GET") // list connections
	r.HandleFunc("/connections/{id}", d.createConnectionHandler).Methods("POST") //create connection
	r.HandleFunc("/connections/{id}", d.connectionInfoHandler).Methods("GET") // get info for connection
	r.HandleFunc("/connections/{id}", d.deleteConnectionHandler).Methods("DELETE") // delete connection
	r.HandleFunc("/connections/{id}/{route}", d.queryRouteHandler).Methods("GET") // get from block route
	http.Handle("/", r)
	
	err := http.ListenAndServe(":" + d.Port, nil)
	if err != nil {
		log.Fatalf(err.Error())
	}
}

type LogMsg struct {
    Type string
    Data interface{}
    Id string
}

func BroadcastStreams(ui chan *LogMsg, logger chan *LogMsg){
	for{
		select{
		case l := <-logger:
			j, err := json.Marshal(l)
			if err != nil {
				log.Println("could not broadcast")
				break
			}
			log.Println(string(j))
			logStream.Broadcast <- j
		case l := <-ui:
			j, err := json.Marshal(l)
			if err != nil {
				log.Println("could not broadcast")
				break
			}
			uiStream.Broadcast <- j
		}
	}
}