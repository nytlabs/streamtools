package server

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/nytlabs/streamtools/st/util"
	"github.com/nytlabs/streamtools/st/library"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"time"
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

type Server struct {
	manager *BlockManager
	mu      *sync.Mutex
	log     chan *util.LogMsg
	ui      chan *util.LogMsg
	Port    string
	Domain  string
	Id      string
}

func NewServer() *Server {
	return &Server{
		manager: NewBlockManager(),
		log:     make(chan *util.LogMsg, 10),
		ui:      make(chan *util.LogMsg, 10),
		mu:      &sync.Mutex{},
	}
}

var resourceType = map[string]string{
	"lib": "application/javascript; charset=utf-8",
	"js":  "application/javascript; charset=utf-8",
	"css": "text/css; charset=utf-8",
}

func (s *Server) rootHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	data, _ := Asset("gui/index.html")
	w.Write(data)
}

func (s *Server) staticHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	w.Header().Set("Content-Type", resourceType[vars["type"]])
	data, _ := Asset("gui/static/" + vars["type"] + "/" + vars["file"])
	w.Write(data)
}

func (s *Server) libraryHandler(w http.ResponseWriter, r *http.Request) {
	lib, err := json.Marshal(library.BlockDefs)
	if err != nil {
		s.log <- &util.LogMsg{
			Type: util.CREATE,
			Data: "Could not marshal library.",
			Id:   s.Id,
		}
	}
	s.apiWrap(w, r, 200, lib)
}

func (s *Server) portHandler(w http.ResponseWriter, r *http.Request) {
	p := []byte(fmt.Sprintf(`{"Port": "%s"}`, s.Port))
	s.apiWrap(w, r, 200, p)
}

func (s *Server) domainHandler(w http.ResponseWriter, r *http.Request) {
	p := []byte(fmt.Sprintf(`{"Domain": "%s"}`, s.Domain))
	s.apiWrap(w, r, 200, p)
}

func (s *Server) versionHandler(w http.ResponseWriter, r *http.Request) {
	p := []byte(fmt.Sprintf(`{"Version": "%s"}`, util.VERSION))
	s.apiWrap(w, r, 200, p)
}

// serveLogStream handles websocket connections for the streamtools log.
// It is write-only.
func (s *Server) serveLogStream(w http.ResponseWriter, r *http.Request) {
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
	recv := make(chan string)
	c.readPump(recv)
}

// serveUIStream handles websocket connections for the streamtools ui.
// It is read/write. Upon hearing anything sent from the client it dumps
// the current ST state into the websocket.
func (s *Server) serveUIStream(w http.ResponseWriter, r *http.Request) {
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

	recv := make(chan string)

	go func(r chan string) {
		for {
			select {
			case <-r:
				// emit block configuration on message
				s.mu.Lock()
				for _, v := range s.manager.ListBlocks() {
					out, _ := json.Marshal(struct {
						Type string
						Data interface{}
						Id   string
					}{
						util.LogInfo[util.CREATE],
						v,
						s.Id,
					})
					c.send <- out
				}
				for _, v := range s.manager.ListConnections() {
					out, _ := json.Marshal(struct {
						Type string
						Data interface{}
						Id   string
					}{
						util.LogInfo[util.CREATE],
						v,
						s.Id,
					})
					c.send <- out
				}
				s.mu.Unlock()
			}
		}
	}(recv)
	c.readPump(recv)
}

// importHandler accepts a JSON through POST that updats the state of ST
// It handles naming collisions by modifying the incoming block pattern.
func (s *Server) importHandler(w http.ResponseWriter, r *http.Request) {
	var export struct {
		Blocks      []*BlockInfo
		Connections []*ConnectionInfo
	}
	corrected := make(map[string]string)

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		s.apiWrap(w, r, 500, s.response(err.Error()))
		return
	}

	err = json.Unmarshal(body, &export)
	if err != nil {
		s.apiWrap(w, r, 500, s.response(err.Error()))
		return
	}

	for _, block := range export.Blocks {
		corrected[block.Id] = block.Id
		for s.manager.IdExists(corrected[block.Id]) {
			corrected[block.Id] = block.Id + "_" + s.manager.GetId()
		}
	}

	for _, conn := range export.Connections {
		corrected[conn.Id] = conn.Id
		for s.manager.IdExists(corrected[conn.Id]) {
			corrected[conn.Id] = conn.Id + "_" + s.manager.GetId()
		}
	}

	for _, block := range export.Blocks {
		block.Id = corrected[block.Id]
		eblock, err := s.manager.Create(block)
		if err != nil {
			s.apiWrap(w, r, 500, s.response(err.Error()))
			return
		}

		s.ui <- &util.LogMsg{
			Type: util.CREATE,
			Data: eblock,
			Id:   s.Id,
		}

		s.log <- &util.LogMsg{
			Type: util.CREATE,
			Data: fmt.Sprintf("Block %s", block.Id),
			Id:   s.Id,
		}
	}

	for _, conn := range export.Connections {
		conn.Id = corrected[conn.Id]
		conn.FromId = corrected[conn.FromId]
		conn.ToId = corrected[conn.ToId]
		econn, err := s.manager.Connect(conn)
		if err != nil {
			s.apiWrap(w, r, 500, s.response(err.Error()))
			return
		}

		s.log <- &util.LogMsg{
			Type: util.CREATE,
			Data: fmt.Sprintf("Connection %s", conn.Id),
			Id:   s.Id,
		}

		s.ui <- &util.LogMsg{
			Type: util.CREATE,
			Data: econn,
			Id:   s.Id,
		}
	}

	s.log <- &util.LogMsg{
		Type: util.INFO,
		Data: "Import OK",
		Id:   s.Id,
	}

	s.apiWrap(w, r, 200, s.response("OK"))
}

// exportHandler creates a JSON file representing the current block system.
func (s *Server) exportHandler(w http.ResponseWriter, r *http.Request) {
	export := struct {
		Blocks      []*BlockInfo
		Connections []*ConnectionInfo
	}{
		s.manager.ListBlocks(),
		s.manager.ListConnections(),
	}

	jex, err := json.Marshal(export)
	if err != nil {
		s.apiWrap(w, r, 500, s.response(err.Error()))
		return
	}

	s.apiWrap(w, r, 200, jex)
}

// listBlockHandler retuns a slice of the current blocks operating in the sytem.
func (s *Server) listBlockHandler(w http.ResponseWriter, r *http.Request) {
	blocks, err := json.Marshal(s.manager.ListBlocks())
	if err != nil {
		s.apiWrap(w, r, 500, s.response(err.Error()))
		return
	}
	s.apiWrap(w, r, 200, blocks)
}

// createBlockHandler asks the manager to create a block and then return that block
// if the block has been creates.
func (s *Server) createBlockHandler(w http.ResponseWriter, r *http.Request) {
	var block *BlockInfo

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		s.apiWrap(w, r, 500, s.response(err.Error()))
		return
	}

	err = json.Unmarshal(body, &block)
	if err != nil {
		s.apiWrap(w, r, 500, s.response(err.Error()))
		return
	}

	mblock, err := s.manager.Create(block)

	if err != nil {
		s.apiWrap(w, r, 500, s.response(err.Error()))
		return
	}

	s.ui <- &util.LogMsg{
		Type: util.CREATE,
		Data: mblock,
		Id:   s.Id,
	}

	s.log <- &util.LogMsg{
		Type: util.CREATE,
		Data: fmt.Sprintf("Block %s", mblock.Id),
		Id:   s.Id,
	}

	jblock, err := json.Marshal(mblock)
	if err != nil {
		s.apiWrap(w, r, 500, s.response(err.Error()))
		return
	}

	s.apiWrap(w, r, 200, jblock)
}

// updateBlockHandler updates the coordinates of a block.
// block.id and block.type can't be changes. block.rule is set through sendRoute
func (s *Server) updateBlockHandler(w http.ResponseWriter, r *http.Request) {
	var coord *Coords
	vars := mux.Vars(r)

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		s.apiWrap(w, r, 500, s.response(err.Error()))
		return
	}

	err = json.Unmarshal(body, &coord)
	if err != nil {
		s.apiWrap(w, r, 500, s.response(err.Error()))
		return
	}

	mblock, err := s.manager.UpdateBlock(vars["id"], coord)

	if err != nil {
		s.apiWrap(w, r, 500, s.response(err.Error()))
		return
	}

	jblock, err := json.Marshal(mblock)
	if err != nil {
		s.apiWrap(w, r, 500, s.response(err.Error()))
		return
	}

	s.log <- &util.LogMsg{
		Type: util.UPDATE,
		Data: fmt.Sprintf("Block %s", mblock.Id),
		Id:   s.Id,
	}

	s.ui <- &util.LogMsg{
		Type: util.UPDATE,
		Data: mblock,
		Id:   s.Id,
	}

	s.apiWrap(w, r, 200, jblock)
}

// blockInfoHandler returns a block given an id
func (s *Server) blockInfoHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	conn, err := s.manager.GetBlock(vars["id"])
	if err != nil {
		s.apiWrap(w, r, 500, s.response(err.Error()))
		return
	}

	jconn, err := json.Marshal(conn)
	if err != nil {
		s.apiWrap(w, r, 500, s.response(err.Error()))
		return
	}
	s.apiWrap(w, r, 200, jconn)
}

// deleteBlockHandler asks the block manager to delete a block.
func (s *Server) deleteBlockHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	ids, err := s.manager.DeleteBlock(vars["id"])
	if err != nil {
		s.apiWrap(w, r, 500, s.response(err.Error()))
		return
	}

	for _, v := range ids {
		s.log <- &util.LogMsg{
			Type: util.DELETE,
			Data: fmt.Sprintf("Block %s", v),
			Id:   s.Id,
		}

		s.ui <- &util.LogMsg{
			Type: util.DELETE,
			Data: struct {
				Id string
			}{
				v,
			},
			Id: s.Id,
		}
	}

	s.apiWrap(w, r, 200, s.response("OK"))
}

// sendRouteHandler sends a message to a block's route. (unidirectional)
func (s *Server) sendRouteHandler(w http.ResponseWriter, r *http.Request) {
	var msg interface{}
	vars := mux.Vars(r)

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		s.apiWrap(w, r, 500, s.response(err.Error()))
		return
	}

	err = json.Unmarshal(body, &msg)
	if err != nil {
		s.apiWrap(w, r, 500, s.response(err.Error()))
		return
	}
	err = s.manager.Send(vars["id"], vars["route"], msg)

	if err != nil {
		s.apiWrap(w, r, 500, s.response(err.Error()))
		return
	}

	s.log <- &util.LogMsg{
		Type: util.UPDATE,
		Data: fmt.Sprintf("Block %s", vars["id"]),
		Id:   s.Id,
	}

	s.ui <- &util.LogMsg{
		Type: util.UPDATE,
		Data: struct {
			Id string
		}{
			vars["id"],
		},
		Id: s.Id,
	}

	s.apiWrap(w, r, 200, s.response("OK"))
}

// queryRouteHandler queries a block and returns a msg. (bidirectional)
func (s *Server) queryRouteHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	msg, err := s.manager.Query(vars["id"], vars["route"])
	if err != nil {
		s.apiWrap(w, r, 500, s.response(err.Error()))
		return
	}

	jmsg, err := json.Marshal(msg)
	if err != nil {
		s.apiWrap(w, r, 500, s.response(err.Error()))
		return
	}

	s.log <- &util.LogMsg{
		Type: util.QUERY,
		Data: fmt.Sprintf("Block %s", vars["id"]),
		Id:   s.Id,
	}

	s.ui <- &util.LogMsg{
		Type: util.QUERY,
		Data: struct {
			Id string
		}{
			vars["id"],
		},
		Id: s.Id,
	}

	s.apiWrap(w, r, 200, jmsg)
}

// listConnectionHandler returns a slice of the current connections in streamtools.
func (s *Server) listConnectionHandler(w http.ResponseWriter, r *http.Request) {
	conns, err := json.Marshal(s.manager.ListConnections())
	if err != nil {
		s.apiWrap(w, r, 500, s.response(err.Error()))
		return
	}
	s.apiWrap(w, r, 200, conns)
}

// createConnectHandler creates a connection and returns it.
func (s *Server) createConnectionHandler(w http.ResponseWriter, r *http.Request) {
	var conn *ConnectionInfo

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		s.apiWrap(w, r, 500, s.response(err.Error()))
		return
	}

	err = json.Unmarshal(body, &conn)
	if err != nil {
		s.apiWrap(w, r, 500, s.response(err.Error()))
		return
	}

	mconn, err := s.manager.Connect(conn)

	if err != nil {
		s.apiWrap(w, r, 500, s.response(err.Error()))
		return
	}

	s.log <- &util.LogMsg{
		Type: util.CREATE,
		Data: fmt.Sprintf("Connection %s", mconn.Id),
		Id:   s.Id,
	}

	s.ui <- &util.LogMsg{
		Type: util.CREATE,
		Data: mconn,
		Id:   s.Id,
	}

	jconn, err := json.Marshal(mconn)
	if err != nil {
		s.apiWrap(w, r, 500, s.response(err.Error()))
		return
	}

	s.apiWrap(w, r, 200, jconn)
}

// connectionInfoHandler returns a connection object, given an is.
func (s *Server) connectionInfoHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	conn, err := s.manager.GetConnection(vars["id"])
	if err != nil {
		s.apiWrap(w, r, 500, s.response(err.Error()))
		return
	}

	jconn, err := json.Marshal(conn)
	if err != nil {
		s.apiWrap(w, r, 500, s.response(err.Error()))
		return
	}
	s.apiWrap(w, r, 200, jconn)
}

// response wraps error responses from the daemon in JSON
func (s *Server) response(statusTxt string) []byte {
	response, err := json.Marshal(struct {
		StatusTxt string `json:"daemon"`
	}{
		statusTxt,
	})
	if err != nil {
		response = []byte(fmt.Sprintf(`{"%s":"%s"}`, s.Id, err.Error()))
	}
	return response
}

// apiWrap wraps all HTTP responses with approprite headers, status codes, and logs them.
func (s *Server) apiWrap(w http.ResponseWriter, r *http.Request, statusCode int, data []byte) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	w.Write(data)

	if statusCode == 200 {
		s.log <- &util.LogMsg{
			Type: util.INFO,
			Data: fmt.Sprintf("%d", statusCode) + ": " + r.URL.Path,
			Id:   s.Id,
		}
	} else {
		var err struct {
			DAEMON string
		}
		_ = json.Unmarshal(data, &err)
		s.log <- &util.LogMsg{
			Type: util.ERROR,
			Data: err.DAEMON,
			Id:   s.Id,
		}
	}
}

// deleteConnectionHandler deletes a connection, responds with OK.
func (s *Server) deleteConnectionHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := s.manager.DeleteConnection(vars["id"])
	if err != nil {
		s.apiWrap(w, r, 500, s.response(err.Error()))
		return
	}

	s.log <- &util.LogMsg{
		Type: util.DELETE,
		Data: fmt.Sprintf("Connection %s", id),
		Id:   s.Id,
	}

	s.ui <- &util.LogMsg{
		Type: util.DELETE,
		Data: struct {
			Id string
		}{
			id,
		},
		Id: s.Id,
	}

	s.apiWrap(w, r, 200, s.response("OK"))
}

func (s *Server) Run() {
	go BroadcastStream(s.ui, s.log)
	go logStream.run()
	go uiStream.run()

	r := mux.NewRouter()
	r.HandleFunc("/", s.rootHandler)
	r.HandleFunc("/library", s.libraryHandler)
	r.HandleFunc("/static/{type}/{file}", s.staticHandler)
	r.HandleFunc("/log", s.serveLogStream)
	r.HandleFunc("/ui", s.serveUIStream)
	r.HandleFunc("/port", s.portHandler)
	r.HandleFunc("/domain", s.domainHandler)
	r.HandleFunc("/version", s.versionHandler)
	r.HandleFunc("/import", s.importHandler).Methods("POST")
	r.HandleFunc("/export", s.exportHandler).Methods("GET")
	r.HandleFunc("/blocks", s.listBlockHandler).Methods("GET")                     // list all blocks
	r.HandleFunc("/blocks", s.createBlockHandler).Methods("POST")                  // create block w/o id
	r.HandleFunc("/blocks/{id}", s.blockInfoHandler).Methods("GET")                // get block info
	r.HandleFunc("/blocks/{id}", s.updateBlockHandler).Methods("PUT")              // update block
	r.HandleFunc("/blocks/{id}", s.deleteBlockHandler).Methods("DELETE")           // delete block
	r.HandleFunc("/blocks/{id}/{route}", s.sendRouteHandler).Methods("POST")       // send to block route
	r.HandleFunc("/blocks/{id}/{route}", s.queryRouteHandler).Methods("GET")       // get from block route
	r.HandleFunc("/connections", s.createConnectionHandler).Methods("POST")        // create connection
	r.HandleFunc("/connections", s.listConnectionHandler).Methods("GET")           // list connections
	r.HandleFunc("/connections/{id}", s.connectionInfoHandler).Methods("GET")      // get info for connection
	r.HandleFunc("/connections/{id}", s.deleteConnectionHandler).Methods("DELETE") // delete connection
	r.HandleFunc("/connections/{id}/{route}", s.queryRouteHandler).Methods("GET")  // get from block route
	http.Handle("/", r)

	s.log <- &util.LogMsg{
		Type: util.INFO,
		Data: fmt.Sprintf("Starting Streamtools %s on port %s", util.VERSION, s.Port),
		Id:   s.Id,
	}

	err := http.ListenAndServe(":"+s.Port, nil)
	if err != nil {
		log.Fatalf(err.Error())
	}
}

// BroadcastStream routes logs and block system changes to websocket hubs
// and terminal.
func BroadcastStream(ui chan *util.LogMsg, logger chan *util.LogMsg) {
	var batch []interface{}

	// we batch the logs every 50 ms so we can cut down on the amount
	// of messages we send
	dump := time.NewTicker(50 * time.Millisecond)

	for {
		select {
		case <-dump.C:
			if len(batch) == 0 {
				break
			}

			outBatch := struct {
				Log []interface{}
			}{
				batch,
			}

			joutBatch, err := json.Marshal(outBatch)
			if err != nil {
				log.Println("could not broadcast")
			}

			logStream.Broadcast <- joutBatch
			batch = nil
		case l := <-logger:
			bclog := struct {
				Type string
				Data interface{}
				Id   string
			}{
				util.LogInfo[l.Type],
				l.Data,
				l.Id,
			}

			fmt.Println(fmt.Sprintf("%s [ %s ][ %s ] %s", time.Now().Format(time.Stamp), l.Id, util.LogInfoColor[l.Type], l.Data))
			batch = append(batch, bclog)
		case l := <-ui:
			bclog := struct {
				Type string
				Data interface{}
				Id   string
			}{
				util.LogInfo[l.Type],
				l.Data,
				l.Id,
			}

			j, err := json.Marshal(bclog)
			if err != nil {
				log.Println("could not broadcast")
				break
			}
			uiStream.Broadcast <- j
		}
	}
}
