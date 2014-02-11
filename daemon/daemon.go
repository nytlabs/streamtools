package daemon

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"io/ioutil"
	"log"
	"net/http"
)

const (
	ADD_CHAN = iota
	DEL_CHAN	
)

var logStream = hub{
	Broadcast:   make(chan []byte),
	register:    make(chan *connection),
	unregister:  make(chan *connection),
	connections: make(map[*connection]bool),
}

var uiStream = hub{
	Broadcast:   make(chan []byte),
	register:    make(chan *connection),
	unregister:  make(chan *connection),
	connections: make(map[*connection]bool),
}

type Daemon struct {
	manager *BlockManager
	log     chan *LogMsg
	ui 		chan *LogMsg
	Port    string
}

func NewDaemon() *Daemon {
	return &Daemon{
		manager: NewBlockManager(),
		log:     make(chan *LogMsg),
		ui: 	 make(chan *LogMsg),
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
	var export struct {
		Blocks      []*BlockInfo
		Connections []*ConnectionInfo
	}
	corrected := make(map[string]string)

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		d.apiWrap(w, 500, d.response(err.Error()))
		return
	}

	err = json.Unmarshal(body, &export)
	if err != nil {
		d.apiWrap(w, 500, d.response(err.Error()))
		return
	}

	for _, block := range export.Blocks {
		corrected[block.Id] = block.Id
		for d.manager.IdExists(corrected[block.Id]) {
			corrected[block.Id] = block.Id + "_" + d.manager.GetId()
		}
	}

	for _, conn := range export.Connections {
		if d.manager.IdExists(conn.Id) {
			corrected[conn.Id] = conn.Id + "_" + d.manager.GetId()
		}
	}

	for _, block := range export.Blocks {
		block.Id = corrected[block.Id]
		_, err := d.manager.Create(block)
		if err != nil {
			d.apiWrap(w, 500, d.response(err.Error()))
			return
		}
	}

	for _, conn := range export.Connections {
		conn.Id = corrected[conn.Id]
		conn.FromId = corrected[conn.FromId]
		conn.ToId = corrected[conn.ToId]
		_, err := d.manager.Connect(conn)
		if err != nil {
			d.apiWrap(w, 500, d.response(err.Error()))
			return
		}
	}

	d.apiWrap(w, 200, d.response("OK"))
}

func (d *Daemon) exportHandler(w http.ResponseWriter, r *http.Request) {
	export := struct {
		Blocks      []*BlockInfo
		Connections []*ConnectionInfo
	}{
		d.manager.ListBlocks(),
		d.manager.ListConnections(),
	}

	jex, err := json.Marshal(export)
	if err != nil {
		d.apiWrap(w, 500, d.response(err.Error()))
		return
	}

	d.apiWrap(w, 200, jex)
}

func (d *Daemon) listBlockHandler(w http.ResponseWriter, r *http.Request) {
	blocks, err := json.Marshal(d.manager.ListBlocks())
	if err != nil {
		d.apiWrap(w, 500, d.response(err.Error()))
		return
	}
	d.apiWrap(w, 200, blocks)
}

func (d *Daemon) createBlockHandler(w http.ResponseWriter, r *http.Request) {
	var block *BlockInfo

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		d.apiWrap(w, 500, d.response(err.Error()))
		return
	}

	err = json.Unmarshal(body, &block)
	if err != nil {
		d.apiWrap(w, 500, d.response(err.Error()))
		return
	}

	mblock, err := d.manager.Create(block)

	if err != nil {
		d.apiWrap(w, 500, d.response(err.Error()))
		return
	}

	jblock, err := json.Marshal(mblock)
	if err != nil {
		d.apiWrap(w, 500, d.response(err.Error()))
		return
	}

	d.log <- &LogMsg{
	    Type: "INFO",
	    Data: "BLOCK CREATED",
	    Id: "DAEMON",
	}

	d.ui <- &LogMsg{
		Type: "CREATE",
		Data: jblock,
		Id: "DAEMON",
	}

	d.apiWrap(w, 200, jblock)
}

func (d *Daemon) updateBlockHandler(w http.ResponseWriter, r *http.Request) {
	var coord *Coords
	vars := mux.Vars(r)

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		d.apiWrap(w, 500, d.response(err.Error()))
		return
	}

	err = json.Unmarshal(body, &coord)
	if err != nil {
		d.apiWrap(w, 500, d.response(err.Error()))
		return
	}

	mblock, err := d.manager.UpdateBlock(vars["id"], coord)

	if err != nil {
		d.apiWrap(w, 500, d.response(err.Error()))
		return
	}

	jblock, err := json.Marshal(mblock)
	if err != nil {
		d.apiWrap(w, 500, d.response(err.Error()))
		return
	}

	d.apiWrap(w, 200, jblock)
}

func (d *Daemon) blockInfoHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	conn, err := d.manager.GetBlock(vars["id"])
	if err != nil {
		d.apiWrap(w, 500, d.response(err.Error()))
		return
	}

	jconn, err := json.Marshal(conn)
	if err != nil {
		d.apiWrap(w, 500, d.response(err.Error()))
		return
	}
	d.apiWrap(w, 200, jconn)
}

func (d *Daemon) deleteBlockHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	err := d.manager.DeleteBlock(vars["id"])
	if err != nil {
		d.apiWrap(w, 500, d.response(err.Error()))
		return
	}

	d.ui <- &LogMsg{
		Type: "DELETE",
		Data: struct{
			Id string
		}{
			vars["id"],
		},
		Id: "DAEMON",
	}

	d.apiWrap(w, 200, d.response("OK"))
}

func (d *Daemon) sendRouteHandler(w http.ResponseWriter, r *http.Request) {
	var msg interface{}
	vars := mux.Vars(r)

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		d.apiWrap(w, 500, d.response(err.Error()))
		return
	}

	err = json.Unmarshal(body, &msg)
	if err != nil {
		d.apiWrap(w, 500, d.response(err.Error()))
		return
	}
	err = d.manager.Send(vars["id"], vars["route"], msg)

	if err != nil {
		d.apiWrap(w, 500, d.response(err.Error()))
		return
	}
	d.apiWrap(w, 200, d.response("OK"))
}

func (d *Daemon) queryRouteHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	msg, err := d.manager.Query(vars["id"], vars["route"])
	if err != nil {
		d.apiWrap(w, 500, d.response(err.Error()))
		return
	}

	jmsg, err := json.Marshal(msg)
	if err != nil {
		d.apiWrap(w, 500, d.response(err.Error()))
		return
	}

	d.apiWrap(w, 200, jmsg)
}

func (d *Daemon) listConnectionHandler(w http.ResponseWriter, r *http.Request) {
	conns, err := json.Marshal(d.manager.ListConnections())
	if err != nil {
		d.apiWrap(w, 500, d.response(err.Error()))
		return
	}
	d.apiWrap(w, 200, conns)
}

func (d *Daemon) createConnectionHandler(w http.ResponseWriter, r *http.Request) {
	var conn *ConnectionInfo

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		d.apiWrap(w, 500, d.response(err.Error()))
		return
	}

	err = json.Unmarshal(body, &conn)
	if err != nil {
		d.apiWrap(w, 500, d.response(err.Error()))
		return
	}

	mconn, err := d.manager.Connect(conn)

	if err != nil {
		d.apiWrap(w, 500, d.response(err.Error()))
		return
	}

	jconn, err := json.Marshal(mconn)
	if err != nil {
		d.apiWrap(w, 500, d.response(err.Error()))
		return
	}

	d.apiWrap(w, 200, jconn)
}

func (d *Daemon) connectionInfoHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	conn, err := d.manager.GetConnection(vars["id"])
	if err != nil {
		d.apiWrap(w, 500, d.response(err.Error()))
		return
	}

	jconn, err := json.Marshal(conn)
	if err != nil {
		d.apiWrap(w, 500, d.response(err.Error()))
		return
	}
	d.apiWrap(w, 200, jconn)
}

func (d *Daemon) response(statusTxt string) []byte {
	response, err := json.Marshal(struct {
		StatusTxt string `json:"daemon"`
	}{
		statusTxt,
	})
	if err != nil {
		response = []byte(fmt.Sprintf(`{"daemon":"%s"}`, err.Error()))
	}
	return response
}

func (d *Daemon) apiWrap(w http.ResponseWriter, statusCode int, data []byte) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	w.Write(data)
}

func (d *Daemon) deleteConnectionHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	err := d.manager.DeleteConnection(vars["id"])
	if err != nil {
		d.apiWrap(w, 500, d.response(err.Error()))
		return
	}
	d.apiWrap(w, 200, d.response("OK"))
}

func (d *Daemon) Run() {
	go BroadcastStream(d.ui, d.log)
	go logStream.run()
	go uiStream.run()

	r := mux.NewRouter()
	r.HandleFunc("/", d.rootHandler)
	r.HandleFunc("/static/{file}", d.staticHandler)
	r.HandleFunc("/log", d.serveLogStream)
	r.HandleFunc("/ui", d.serveUIStream)
	r.HandleFunc("/import", d.importHandler).Methods("POST")
	r.HandleFunc("/export", d.exportHandler).Methods("GET")
	r.HandleFunc("/blocks", d.listBlockHandler).Methods("GET")                     // list all blocks
	r.HandleFunc("/blocks", d.createBlockHandler).Methods("POST")                  // create block w/o id
	r.HandleFunc("/blocks/{id}", d.blockInfoHandler).Methods("GET")                // get block info
	r.HandleFunc("/blocks/{id}", d.updateBlockHandler).Methods("PUT")           	// update block
	r.HandleFunc("/blocks/{id}", d.deleteBlockHandler).Methods("DELETE")           // delete block
	r.HandleFunc("/blocks/{id}/{route}", d.sendRouteHandler).Methods("POST")       // send to block route
	r.HandleFunc("/blocks/{id}/{route}", d.queryRouteHandler).Methods("GET")       // get from block route
	r.HandleFunc("/connections", d.createConnectionHandler).Methods("POST")        // create connection
	r.HandleFunc("/connections", d.listConnectionHandler).Methods("GET")           // list connections
	r.HandleFunc("/connections/{id}", d.connectionInfoHandler).Methods("GET")      // get info for connection
	r.HandleFunc("/connections/{id}", d.deleteConnectionHandler).Methods("DELETE") // delete connection
	r.HandleFunc("/connections/{id}/{route}", d.queryRouteHandler).Methods("GET")  // get from block route
	http.Handle("/", r)

	err := http.ListenAndServe(":"+d.Port, nil)
	if err != nil {
		log.Fatalf(err.Error())
	}
}

type LogMsg struct {
	Type string
	Data interface{}
	Id   string
}

func BroadcastStream(ui chan *LogMsg, logger chan *LogMsg) {
	for {
		select {
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
