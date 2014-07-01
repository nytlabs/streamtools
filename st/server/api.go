package server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"runtime/pprof"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/nytlabs/streamtools/st/blocks"
	"github.com/nytlabs/streamtools/st/library"
	"github.com/nytlabs/streamtools/st/loghub"
	"github.com/nytlabs/streamtools/st/util"
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
	Port    string
	Domain  string
	Id      string
}

func NewServer() *Server {
	return &Server{
		manager: NewBlockManager(),
	}
}

var resourceType = map[string]string{
	"html": "text/html; charset=utf-8",
	"lib":  "application/javascript; charset=utf-8",
	"js":   "application/javascript; charset=utf-8",
	"json": "application/javascript; charset=utf-8",
	"css":  "text/css; charset=utf-8",
}

func (s *Server) rootHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	data, _ := Asset("gui/index.html")
	w.Write(data)
}

func (s *Server) exampleHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	w.Header().Set("Content-Type", resourceType[vars["type"]])
	data, _ := Asset("examples/" + vars["file"])
	s.apiWrap(w, r, 200, data)
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
		loghub.Log <- &loghub.LogMsg{
			Type: loghub.CREATE,
			Data: "Could not marshal library.",
			Id:   s.Id,
		}
	}
	s.apiWrap(w, r, 200, lib)
}

func (s *Server) versionHandler(w http.ResponseWriter, r *http.Request) {
	p := []byte(fmt.Sprintf(`{"Version": "%s"}`, util.VERSION))
	s.apiWrap(w, r, 200, p)
}

func (s *Server) clearHandler(w http.ResponseWriter, r *http.Request) {
	s.manager.Mu.Lock()
	defer s.manager.Mu.Unlock()

	conns := s.manager.ListConnections()
	for _, v := range conns {
		id, err := s.manager.DeleteConnection(v.Id)
		if err != nil {
			loghub.Log <- &loghub.LogMsg{
				Type: loghub.DELETE,
				Data: err.Error(),
				Id:   s.Id,
			}
		}

		loghub.UI <- &loghub.LogMsg{
			Type: loghub.DELETE,
			Data: struct {
				Id string
			}{
				id,
			},
			Id: s.Id,
		}
	}

	blocks := s.manager.ListBlocks()
	for _, v := range blocks {
		ids, err := s.manager.DeleteBlock(v.Id)
		if err != nil {
			loghub.Log <- &loghub.LogMsg{
				Type: loghub.DELETE,
				Data: err.Error(),
				Id:   s.Id,
			}
			continue
		}

		for _, id := range ids {
			loghub.Log <- &loghub.LogMsg{
				Type: loghub.DELETE,
				Data: fmt.Sprintf("Block %s", id),
				Id:   s.Id,
			}

			loghub.UI <- &loghub.LogMsg{
				Type: loghub.DELETE,
				Data: struct {
					Id string
				}{
					id,
				},
				Id: s.Id,
			}
		}
	}

	loghub.Log <- &loghub.LogMsg{
		Type: loghub.INFO,
		Data: fmt.Sprintf("Go routines: %d", runtime.NumGoroutine()),
		Id:   s.Id,
	}

	s.apiWrap(w, r, 200, s.response("OK"))
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
		//log.Println(err)
		return
	}
	c := &connection{send: make(chan []byte, 256), ws: ws, Hub: logStream}
	c.Hub.register <- c
	go c.writePump()
	recv := make(chan string)
	c.readPump(recv)
}

func (s *Server) websocketHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	blockId, ok := vars["id"]
	if !ok {
		s.apiWrap(w, r, 500, s.response("must specify block ID to connect"))
		return
	}

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
		s.apiWrap(w, r, 500, s.response("Not a websocket handshake"))
		return
	} else if err != nil {
		//log.Println(err)
		return
	}
	c := &connection{send: make(chan []byte, 256), ws: ws}

	s.manager.Mu.Lock()
	blockChan, connId, err := s.manager.GetSocket(vars["id"])
	s.manager.Mu.Unlock()

	if err != nil {
		s.apiWrap(w, r, 500, s.response(err.Error()))
		return
	}

	ticker := time.NewTicker((10 * time.Second * 9) / 10)
	go func(c *connection, bChan chan *blocks.Msg, cId string, bId string) {
		defer func() {
			s.manager.Mu.Lock()
			_ = s.manager.DeleteSocket(bId, cId)
			s.manager.Mu.Unlock()
			ticker.Stop()
			c.ws.Close()
		}()

		// start the writePump
		go c.writePump()

		for {
			select {
			case msg := <-bChan:
				message, err := json.Marshal(msg.Msg)
				if err != nil {
					s.apiWrap(w, r, 500, s.response("Attempted to send non JSON encoded message on websocket"))
					return
				}
				select {
				case c.send <- message:
				default:
					loghub.Log <- &loghub.LogMsg{
						Type: loghub.ERROR,
						Data: "websocket send is blocked! Exiting.",
						Id:   s.Id,
					}
					return
				}
			}
		}
	}(c, blockChan, connId, blockId)
}

func (s *Server) streamHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	blockId, ok := vars["id"]
	if !ok {
		s.apiWrap(w, r, 500, s.response("must specify block ID to connect"))
		return
	}
	s.manager.Mu.Lock()
	blockChan, connId, err := s.manager.GetSocket(blockId)
	s.manager.Mu.Unlock()

	if err != nil {
		s.apiWrap(w, r, 500, s.response(err.Error()))
		return
	}

	for {
		msg := <-blockChan
		message, _ := json.Marshal(msg.Msg)
		_, err := w.Write(message)
		_, err = w.Write([]byte("\r\n"))
		if err != nil {
			s.manager.Mu.Lock()
			s.manager.DeleteSocket(blockId, connId)
			s.manager.Mu.Unlock()
			break
		}
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}
	}
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
		//log.Println(err)
		return
	}
	c := &connection{send: make(chan []byte, 256), ws: ws, Hub: uiStream}
	c.Hub.register <- c
	go c.writePump()

	recv := make(chan string)

	go func(r chan string) {
		for {
			select {
			case msgWS := <-r:
				var msg map[string]interface{}
				err := json.Unmarshal([]byte(msgWS), &msg)
				if err != nil {
					loghub.Log <- &loghub.LogMsg{
						Type: loghub.ERROR,
						Data: err.Error(),
						Id:   s.Id,
					}
					break
				}

				_, ok := msg["action"]
				if !ok {
					loghub.Log <- &loghub.LogMsg{
						Type: loghub.ERROR,
						Data: "could not understand websocket request",
						Id:   s.Id,
					}
					break
				}

				actStr, ok := msg["action"].(string)
				if !ok {
					loghub.Log <- &loghub.LogMsg{
						Type: loghub.ERROR,
						Data: "could not understand websocket request",
						Id:   s.Id,
					}
					break
				}

				switch actStr {
				case "quit":
					// so that we don't send to a closed websocket.
					return
				case "export":
					// emit block configuration on message
					s.manager.Mu.Lock()
					for _, v := range s.manager.ListBlocks() {
						out, _ := json.Marshal(struct {
							Type string
							Data interface{}
							Id   string
						}{
							loghub.LogInfo[loghub.CREATE],
							v,
							s.Id,
						})
						if err := c.write(websocket.TextMessage, out); err != nil {
							return
						}
					}
					for _, v := range s.manager.ListConnections() {
						out, _ := json.Marshal(struct {
							Type string
							Data interface{}
							Id   string
						}{
							loghub.LogInfo[loghub.CREATE],
							v,
							s.Id,
						})
						if err := c.write(websocket.TextMessage, out); err != nil {
							return
						}
					}
					s.manager.Mu.Unlock()
				case "rule":
					_, ok := msg["id"]
					if !ok {
						break
					}
					idStr, ok := msg["id"].(string)
					if !ok {
						break
					}
					s.manager.Mu.Lock()
					b, _ := s.manager.GetBlock(idStr)
					s.manager.Mu.Unlock()
					out, _ := json.Marshal(struct {
						Type string
						Data interface{}
						Id   string
					}{
						loghub.LogInfo[loghub.UPDATE_RULE],
						b,
						s.Id,
					})
					if err := c.write(websocket.TextMessage, out); err != nil {
						return
					}
				}
			}
		}
	}(recv)
	c.readPump(recv)
}

// Handles OPTIONS requests for cross-domain pattern importing (see streamtools-tutorials)
func (s *Server) optionsHandler(w http.ResponseWriter, r *http.Request) {
	s.apiWrap(w, r, 200, s.response("OK"))
}

func (s *Server) ImportFile(filename string) {
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		loghub.Log <- &loghub.LogMsg{
			Type: loghub.ERROR,
			Data: err.Error(),
			Id:   s.Id,
		}
	}

	err = s.importJSON(b)
	if err != nil {
		loghub.Log <- &loghub.LogMsg{
			Type: loghub.ERROR,
			Data: err.Error(),
			Id:   s.Id,
		}
	}
}

func (s *Server) importJSON(body []byte) error {
	s.manager.Mu.Lock()
	defer s.manager.Mu.Unlock()

	var export struct {
		Blocks      []*BlockInfo
		Connections []*ConnectionInfo
	}

	corrected := make(map[string]string)

	err := json.Unmarshal(body, &export)
	if err != nil {
		return err
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
			return err
		}

		loghub.UI <- &loghub.LogMsg{
			Type: loghub.CREATE,
			Data: eblock,
			Id:   s.Id,
		}

		loghub.Log <- &loghub.LogMsg{
			Type: loghub.CREATE,
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
			return err
		}

		loghub.Log <- &loghub.LogMsg{
			Type: loghub.CREATE,
			Data: fmt.Sprintf("Connection %s", conn.Id),
			Id:   s.Id,
		}

		loghub.UI <- &loghub.LogMsg{
			Type: loghub.CREATE,
			Data: econn,
			Id:   s.Id,
		}
	}

	loghub.Log <- &loghub.LogMsg{
		Type: loghub.INFO,
		Data: "Import OK",
		Id:   s.Id,
	}
	return nil
}

// importHandler accepts a JSON through POST that updats the state of ST
// It handles naming collisions by modifying the incoming block pattern.
func (s *Server) importHandler(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		s.apiWrap(w, r, 500, s.response(err.Error()))
		return
	}

	err = s.importJSON(body)
	if err != nil {
		s.apiWrap(w, r, 500, s.response(err.Error()))
		return
	}

	s.apiWrap(w, r, 200, s.response("OK"))
}

// exportHandler creates a JSON file representing the current block system.
func (s *Server) exportHandler(w http.ResponseWriter, r *http.Request) {
	s.manager.Mu.Lock()
	defer s.manager.Mu.Unlock()

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
	s.manager.Mu.Lock()
	defer s.manager.Mu.Unlock()

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
	s.manager.Mu.Lock()
	defer s.manager.Mu.Unlock()

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

	loghub.UI <- &loghub.LogMsg{
		Type: loghub.CREATE,
		Data: mblock,
		Id:   s.Id,
	}

	loghub.Log <- &loghub.LogMsg{
		Type: loghub.CREATE,
		Data: fmt.Sprintf("Block %s", mblock.Id),
		Id:   s.Id,
	}

	loghub.Log <- &loghub.LogMsg{
		Type: loghub.INFO,
		Data: fmt.Sprintf("Go routines: %d", runtime.NumGoroutine()),
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
	s.manager.Mu.Lock()
	defer s.manager.Mu.Unlock()

	var update map[string]interface{}

	vars := mux.Vars(r)
	blockId := vars["id"]

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		s.apiWrap(w, r, 500, s.response(err.Error()))
		return
	}

	err = json.Unmarshal(body, &update)
	if err != nil {
		s.apiWrap(w, r, 500, s.response(err.Error()))
		return
	}

	if _, ok := update["X"]; ok {
		c := &Coords{
			X: update["X"].(float64),
			Y: update["Y"].(float64),
		}

		mblock, err := s.manager.UpdateBlockPosition(blockId, c)

		if err != nil {
			s.apiWrap(w, r, 500, s.response(err.Error()))
			return
		}

		loghub.Log <- &loghub.LogMsg{
			Type: loghub.UPDATE,
			Data: fmt.Sprintf("Block %s", mblock.Id),
			Id:   s.Id,
		}

		loghub.UI <- &loghub.LogMsg{
			Type: loghub.UPDATE_POSITION,
			Data: mblock,
			Id:   s.Id,
		}
	}

	if _, ok := update["Id"]; ok {
		mblock, mconnections, err := s.manager.UpdateBlockId(blockId, update["Id"].(string))
		if err != nil {
			s.apiWrap(w, r, 500, s.response(err.Error()))
			return
		}

		blockId = mblock.Id

		loghub.UI <- &loghub.LogMsg{
			Type: loghub.DELETE,
			Data: struct {
				Id string
			}{
				vars["id"],
			},
			Id: s.Id,
		}

		loghub.UI <- &loghub.LogMsg{
			Type: loghub.CREATE,
			Data: mblock,
			Id:   s.Id,
		}

		for _, c := range mconnections {
			loghub.UI <- &loghub.LogMsg{
				Type: loghub.DELETE,
				Data: struct {
					Id string
				}{
					c.Id,
				},
				Id: s.Id,
			}

			loghub.UI <- &loghub.LogMsg{
				Type: loghub.CREATE,
				Data: c,
				Id:   s.Id,
			}
		}
	}

	block, err := s.manager.GetBlock(blockId)
	if err != nil {
		s.apiWrap(w, r, 500, s.response(err.Error()))
		return
	}

	jblock, err := json.Marshal(block)
	if err != nil {
		s.apiWrap(w, r, 500, s.response(err.Error()))
		return
	}

	s.apiWrap(w, r, 200, jblock)
}

// blockInfoHandler returns a block given an id
func (s *Server) blockInfoHandler(w http.ResponseWriter, r *http.Request) {
	s.manager.Mu.Lock()
	defer s.manager.Mu.Unlock()

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
	s.manager.Mu.Lock()
	defer s.manager.Mu.Unlock()

	vars := mux.Vars(r)
	ids, err := s.manager.DeleteBlock(vars["id"])
	if err != nil {
		s.apiWrap(w, r, 500, s.response(err.Error()))
		return
	}

	for _, v := range ids {
		loghub.Log <- &loghub.LogMsg{
			Type: loghub.DELETE,
			Data: fmt.Sprintf("Block %s", v),
			Id:   s.Id,
		}

		loghub.UI <- &loghub.LogMsg{
			Type: loghub.DELETE,
			Data: struct {
				Id string
			}{
				v,
			},
			Id: s.Id,
		}
	}

	loghub.Log <- &loghub.LogMsg{
		Type: loghub.INFO,
		Data: fmt.Sprintf("Go routines: %d", runtime.NumGoroutine()),
		Id:   s.Id,
	}

	s.apiWrap(w, r, 200, s.response("OK"))
}

// sendRouteHandler sends a message to a block's route. (unidirectional)
func (s *Server) sendRouteHandler(w http.ResponseWriter, r *http.Request) {
	s.manager.Mu.Lock()
	defer s.manager.Mu.Unlock()

	var msg interface{}
	vars := mux.Vars(r)

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		s.apiWrap(w, r, 500, s.response(err.Error()))
		return
	}

	err = json.Unmarshal(body, &msg)
	if err != nil {
		msg = map[string]interface{}{
			"data": string(body),
		}
	}
	err = s.manager.Send(vars["id"], vars["route"], msg)

	if err != nil {
		s.apiWrap(w, r, 500, s.response(err.Error()))
		return
	}

	loghub.Log <- &loghub.LogMsg{
		Type: loghub.UPDATE,
		Data: fmt.Sprintf("Block %s", vars["id"]),
		Id:   s.Id,
	}

	/*b, err := s.manager.GetBlock(vars["id"])
	if err != nil {
		s.apiWrap(w, r, 500, s.response(err.Error()))
	}

	loghub.UI <- &loghub.LogMsg{
		Type: loghub.UPDATE,
		Data: b,
		Id: s.Id,
	}*/

	s.apiWrap(w, r, 200, s.response("OK"))
}

// queryRouteHandler queries a block and returns a msg. (bidirectional)
func (s *Server) queryBlockHandler(w http.ResponseWriter, r *http.Request) {
	s.manager.Mu.Lock()
	defer s.manager.Mu.Unlock()

	vars := mux.Vars(r)
	u, err := url.Parse(r.RequestURI)
	if err != nil {
		s.apiWrap(w, r, 500, s.response(err.Error()))
		return
	}
	params := u.Query()
	var msg interface{}
	if len(params) > 0 {
		msg, err = s.manager.QueryParamBlock(vars["id"], vars["route"], params)
		if err != nil {
			s.apiWrap(w, r, 500, s.response(err.Error()))
			return
		}

	} else {

		msg, err = s.manager.QueryBlock(vars["id"], vars["route"])
		if err != nil {
			s.apiWrap(w, r, 500, s.response(err.Error()))
			return
		}
	}

	jmsg, err := json.Marshal(msg)
	if err != nil {
		s.apiWrap(w, r, 500, s.response(err.Error()))
		return
	}

	loghub.Log <- &loghub.LogMsg{
		Type: loghub.QUERY,
		Data: fmt.Sprintf("Block %s", vars["id"]),
		Id:   s.Id,
	}

	loghub.UI <- &loghub.LogMsg{
		Type: loghub.QUERY,
		Data: struct {
			Id string
		}{
			vars["id"],
		},
		Id: s.Id,
	}

	s.apiWrap(w, r, 200, jmsg)
}

func (s *Server) queryConnectionHandler(w http.ResponseWriter, r *http.Request) {
	s.manager.Mu.Lock()
	defer s.manager.Mu.Unlock()

	vars := mux.Vars(r)

	msg, err := s.manager.QueryConnection(vars["id"], vars["route"])
	if err != nil {
		s.apiWrap(w, r, 500, s.response(err.Error()))
		return
	}

	jmsg, err := json.Marshal(msg)
	if err != nil {
		s.apiWrap(w, r, 500, s.response(err.Error()))
		return
	}

	loghub.Log <- &loghub.LogMsg{
		Type: loghub.QUERY,
		Data: fmt.Sprintf("Connection %s", vars["id"]),
		Id:   s.Id,
	}

	loghub.UI <- &loghub.LogMsg{
		Type: loghub.QUERY,
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
	s.manager.Mu.Lock()
	defer s.manager.Mu.Unlock()

	conns, err := json.Marshal(s.manager.ListConnections())
	if err != nil {
		s.apiWrap(w, r, 500, s.response(err.Error()))
		return
	}
	s.apiWrap(w, r, 200, conns)
}

// createConnectHandler creates a connection and returns it.
func (s *Server) createConnectionHandler(w http.ResponseWriter, r *http.Request) {
	s.manager.Mu.Lock()
	defer s.manager.Mu.Unlock()

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

	loghub.Log <- &loghub.LogMsg{
		Type: loghub.CREATE,
		Data: fmt.Sprintf("Connection %s", mconn.Id),
		Id:   s.Id,
	}

	loghub.UI <- &loghub.LogMsg{
		Type: loghub.CREATE,
		Data: mconn,
		Id:   s.Id,
	}

	loghub.Log <- &loghub.LogMsg{
		Type: loghub.INFO,
		Data: fmt.Sprintf("Go routines: %d", runtime.NumGoroutine()),
		Id:   s.Id,
	}

	jconn, err := json.Marshal(mconn)
	if err != nil {
		s.apiWrap(w, r, 500, s.response(err.Error()))
		return
	}

	s.apiWrap(w, r, 200, jconn)
}

func (s *Server) topHandler(w http.ResponseWriter, r *http.Request) {
	pprof.Lookup("goroutine").WriteTo(os.Stdout, 1)
	s.apiWrap(w, r, 200, s.response("OK"))
}

func (s *Server) statusHandler(w http.ResponseWriter, r *http.Request) {

	s.manager.Mu.Lock()
	defer s.manager.Mu.Unlock()

	export := struct {
		Blocks []string
	}{
		s.manager.StatusBlocks(),
	}

	jex, err := json.Marshal(export)
	if err != nil {
		s.apiWrap(w, r, 500, s.response(err.Error()))
		return
	}

	s.apiWrap(w, r, 200, jex)
}

func (s *Server) profStartHandler(w http.ResponseWriter, r *http.Request) {
	f, err := os.Create("streamtools.prof")
	if err != nil {
		log.Fatal(err)
	}
	pprof.StartCPUProfile(f)

	s.apiWrap(w, r, 200, s.response("OK"))
}

func (s *Server) profStopHandler(w http.ResponseWriter, r *http.Request) {
	pprof.StopCPUProfile()
	s.apiWrap(w, r, 200, s.response("OK"))
}

// connectionInfoHandler returns a connection object, given an is.
func (s *Server) connectionInfoHandler(w http.ResponseWriter, r *http.Request) {
	s.manager.Mu.Lock()
	defer s.manager.Mu.Unlock()

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
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
	w.Header().Set("Access-Control-Allow-Methods", "GET, PUT, POST, DELETE, OPTIONS")
	w.WriteHeader(statusCode)
	w.Write(data)

	if statusCode == 200 {
		loghub.Log <- &loghub.LogMsg{
			Type: loghub.INFO,
			Data: fmt.Sprintf("%d", statusCode) + ": " + r.URL.Path,
			Id:   s.Id,
		}
	} else {
		var err struct {
			DAEMON string
		}
		_ = json.Unmarshal(data, &err)
		loghub.Log <- &loghub.LogMsg{
			Type: loghub.ERROR,
			Data: err.DAEMON,
			Id:   s.Id,
		}
	}
}

// deleteConnectionHandler deletes a connection, responds with OK.
func (s *Server) deleteConnectionHandler(w http.ResponseWriter, r *http.Request) {
	s.manager.Mu.Lock()
	defer s.manager.Mu.Unlock()

	vars := mux.Vars(r)
	id, err := s.manager.DeleteConnection(vars["id"])
	if err != nil {
		s.apiWrap(w, r, 500, s.response(err.Error()))
		return
	}

	loghub.Log <- &loghub.LogMsg{
		Type: loghub.DELETE,
		Data: fmt.Sprintf("Connection %s", id),
		Id:   s.Id,
	}

	loghub.UI <- &loghub.LogMsg{
		Type: loghub.DELETE,
		Data: struct {
			Id string
		}{
			id,
		},
		Id: s.Id,
	}

	loghub.Log <- &loghub.LogMsg{
		Type: loghub.INFO,
		Data: fmt.Sprintf("Go routines: %d", runtime.NumGoroutine()),
		Id:   s.Id,
	}

	s.apiWrap(w, r, 200, s.response("OK"))
}

func (s *Server) Run() {
	go logStream.run()
	go uiStream.run()

	loghub.AddLog <- logStream.Broadcast
	loghub.AddUI <- uiStream.Broadcast

	r := mux.NewRouter()
	r.StrictSlash(true)
	r.HandleFunc("/", s.rootHandler)
	r.HandleFunc("/library", s.libraryHandler)
	r.HandleFunc("/static/{type}/{file}", s.staticHandler)
	r.HandleFunc("/log", s.serveLogStream)
	r.HandleFunc("/ui", s.serveUIStream)
	r.HandleFunc("/version", s.versionHandler)
	r.HandleFunc("/top", s.topHandler)
	r.HandleFunc("/examples/{file}", s.exampleHandler)
	r.HandleFunc("/status", s.statusHandler)
	r.HandleFunc("/profstart", s.profStartHandler)
	r.HandleFunc("/profstop", s.profStopHandler)
	r.HandleFunc("/clear", s.clearHandler).Methods("GET")
	r.HandleFunc("/import", s.importHandler).Methods("POST")
	r.HandleFunc("/import", s.optionsHandler).Methods("OPTIONS")
	r.HandleFunc("/export", s.exportHandler).Methods("GET")
	r.HandleFunc("/blocks", s.listBlockHandler).Methods("GET")                         // list all blocks
	r.HandleFunc("/blocks", s.createBlockHandler).Methods("POST")                      // create block w/o id
	r.HandleFunc("/blocks", s.optionsHandler).Methods("OPTIONS")                       // allow cross-domain
	r.HandleFunc("/blocks/{id}", s.blockInfoHandler).Methods("GET")                    // get block info
	r.HandleFunc("/blocks/{id}", s.updateBlockHandler).Methods("PUT")                  // update block
	r.HandleFunc("/blocks/{id}", s.deleteBlockHandler).Methods("DELETE")               // delete block
	r.HandleFunc("/blocks/{id}/{route}", s.sendRouteHandler).Methods("POST")           // send to block route
	r.HandleFunc("/blocks/{id}/{route}", s.queryBlockHandler).Methods("GET")           // get from block route
	r.HandleFunc("/blocks/{id}/{route}", s.optionsHandler).Methods("OPTIONS")          // allow cross-domain
	r.HandleFunc("/ws/{id}", s.websocketHandler).Methods("GET")                        // websocket handler
	r.HandleFunc("/stream/{id}", s.streamHandler).Methods("GET")                       // http stream handler
	r.HandleFunc("/connections", s.createConnectionHandler).Methods("POST")            // create connection
	r.HandleFunc("/connections", s.optionsHandler).Methods("OPTIONS")                  // allow cross-domain
	r.HandleFunc("/connections", s.listConnectionHandler).Methods("GET")               // list connections
	r.HandleFunc("/connections/{id}", s.connectionInfoHandler).Methods("GET")          // get info for connection
	r.HandleFunc("/connections/{id}", s.deleteConnectionHandler).Methods("DELETE")     // delete connection
	r.HandleFunc("/connections/{id}/{route}", s.queryConnectionHandler).Methods("GET") // get from block route
	http.Handle("/", r)

	loghub.Log <- &loghub.LogMsg{
		Type: loghub.INFO,
		Data: fmt.Sprintf("Starting Streamtools %s on port %s", util.VERSION, s.Port),
		Id:   s.Id,
	}

	err := http.ListenAndServe(":"+s.Port, nil)
	if err != nil {
		log.Fatalf(err.Error())
	}
}
