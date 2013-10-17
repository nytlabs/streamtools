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
	idChan chan string
	port   = flag.String("port", "7070", "stream tools port")
)

type query struct {
	r            *http.Request
	responseChan chan *simplejson.Json
}

type hub struct {
	connectionMap map[string]Block
	blockMap      map[string]Block
}

type routeResponse struct {
	msg          *simplejson.Json
	responseChan chan *simplejson.Json
}

func (self *hub) rootHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "hello! this is streamtools")
	fmt.Fprintln(w, "ID: BlockType")
	for id, block := range self.blockMap {
		fmt.Fprintln(w, id+":", block.getBlockType())
	}
}

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

/*func (self *hub) queryHandler(w http.ResponseWriter, r *http.Request) {
	id := strings.Split(r.URL.Path, "/")[2]
	log.Println("sending query to", id)
	// get the relevant block's query channel
	queryChan := self.blockMap[id].getQueryChan()
	responseChan := make(chan *simplejson.Json)
	// submit the query
	queryChan <- query{r, responseChan}
	// wait for the response
	response := <-responseChan
	out, err := response.MarshalJSON()
	if err != nil {
		log.Println(err.Error())
	}
	fmt.Fprint(w, string(out))
}*/

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

func (self *hub) CreateConnection(from string, to string) {
	conn := NewBlock("connection")
	conn.setInChan(self.blockMap[from].getOutChan())
	conn.setOutChan(self.blockMap[to].getInChan())
	self.connectionMap[conn.getID()] = conn
	go conn.blockRoutine()
}

func (self *hub) CreateBlock(blockType string, id string) {
	block := NewBlock(blockType)
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
