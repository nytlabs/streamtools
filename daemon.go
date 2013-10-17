package streamtools

import (
	"flag"
	"fmt"
	"github.com/bitly/go-simplejson"
	"log"
	"net/http"
	"strings"
)

var (
	// channel that returns the next ID
	idChan chan string
	// port that streamtools reuns on
	port = flag.String("port", "7070", "stream tools port")
)

// hub keeps track of all the blocks and connections
type hub struct {
	connectionMap map[string]Block
	blockMap      map[string]Block
}

// routeResponse is passed into a block to query via established handlers
type routeResponse struct {
	msg          *simplejson.Json
	responseChan chan *simplejson.Json
}

// HANDLERS

// The rootHandler returns information about the whole system
func (self *hub) rootHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "hello! this is streamtools")
	fmt.Fprintln(w, "ID: BlockType")
	for id, block := range self.blockMap {
		fmt.Fprintln(w, id+":", block.getBlockType())
	}
}

// The createHandler creates new blocks
func (self *hub) createHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Println("could not parse form on /create")
	}
	if blockType, ok := r.Form["blockType"]; ok {

		var id string
		if blockId, ok := r.Form["id"]; ok {
			id = blockId[0]
		} else {
			id = <-idChan
		}
		self.CreateBlock(blockType[0], id)

	} else {
		log.Println("no blocktype specified")
	}
}

// The connectHandler connects together two blocks
func (self *hub) connectHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Println("could not parse form on /connect")
	}
	from := r.Form["from"][0]
	to := r.Form["to"][0]
	log.Println("connecting", from, "to", to)
	self.CreateConnection(from, to)
}

// The routeHandler deals with any incoming message sent to an arbitrary block endpoint
func (self *hub) routeHandler(w http.ResponseWriter, r *http.Request) {
	id := strings.Split(r.URL.Path, "/")[2]
	route := strings.Split(r.URL.Path, "/")[3]

	err := r.ParseForm()
	var respData string
	for k, _ := range r.Form {
		respData = k
	}
	msg, err := simplejson.NewJson([]byte(respData))
	if err != nil {
		msg = nil
	}
	responseChan := make(chan *simplejson.Json)
	blockRouteChan := self.blockMap[id].getRouteChan(route)
	blockRouteChan <- routeResponse{
		msg:          msg,
		responseChan: responseChan,
	}
	blockMsg := <-responseChan
	out, err := blockMsg.MarshalJSON()
	if err != nil {
		log.Println(err.Error())
	}

	fmt.Fprint(w, string(out))
}

func (self *hub) libraryHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, libraryBlob)
}

func (self *hub) CreateConnection(from string, to string) {
	conn := library["connection"].blockFactory()
	conn.initOutChans()
	id := <-idChan
	conn.setID(id)

	fromChan := self.blockMap[from].createOutChan(conn.getID())
	conn.setInChan(fromChan)

	toChan := self.blockMap[to].getInChan()
	conn.setOutChan(to, toChan)

	self.connectionMap[conn.getID()] = conn
	go conn.blockRoutine()
}

func (self *hub) CreateBlock(blockType string, id string) {
	blockTemplate, ok := library[blockType]
	if !ok {
		log.Fatal("couldn't find block", blockType)
	}
	block := blockTemplate.blockFactory()
	block.initOutChans()

	block.setID(id)
	self.blockMap[id] = block

	routeNames := block.getRoutes()
	for _, routeName := range routeNames {
		http.HandleFunc("/blocks/"+block.getID()+"/"+routeName, self.routeHandler)
	}

	go block.blockRoutine()
}

func (self *hub) Run() {

	// start the ID Service
	idChan = make(chan string)
	go IDService(idChan)

	// start the library service
	buildLibrary()

	// initialise the connection and block maps
	self.connectionMap = make(map[string]Block)
	self.blockMap = make(map[string]Block)

	// instantiate the base handlers
	http.HandleFunc("/", self.rootHandler)
	http.HandleFunc("/create", self.createHandler)
	http.HandleFunc("/connect", self.connectHandler)
	http.HandleFunc("/library", self.libraryHandler)

	// start the http server
	log.Println("starting stream tools on port", *port)
	err := http.ListenAndServe(":"+*port, nil)
	if err != nil {
		log.Fatalf(err.Error())
	}
}

func Run() {
	h := hub{}
	h.Run()
}
