package streamtools

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"
)

var (
	idChan chan string
	port   = flag.String("port", "7070", "stream tools port")
)

type query struct {
	w http.ResponseWriter
	r *http.Request
}

type hub struct {
	connectionMap map[string]Block
	blockMap      map[string]Block
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
		self.CreateBlock(blockType[0])
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
	self.CreateConnection(from, to)
}

func (self *hub) queryHandler(w http.ResponseWriter, r *http.Request) {
	id := strings.Split(r.URL.Path, "/")[1]
	log.Println("sending query to", id)
	// get the relevant block's query channel
	queryChan := self.blockMap[id].getQueryChan()
	// submit the query
	queryChan <- query{w, r}
}

func (self *hub) CreateConnection(from string, to string) {
	conn := NewBlock("connection")
	conn.setInChan(self.blockMap[from].getOutChan())
	conn.setOutChan(self.blockMap[to].getInChan())
	self.connectionMap[conn.getID()] = conn
}

func (self *hub) CreateBlock(blockType string) {
	block := NewBlock(blockType)
	self.blockMap[block.getID()] = block
	http.HandleFunc("/blocks/"+block.getID()+"/query", self.queryHandler)
	go block.blockRoutine()
}

func (self *hub) Run() {
	idChan = make(chan string)
	go IDService(idChan)
	buildLibrary()

	self.connectionMap = make(map[string]Block)
	self.blockMap = make(map[string]Block)

	http.HandleFunc("/", self.rootHandler)
	http.HandleFunc("/create", self.createHandler)
	http.HandleFunc("/connect", self.connectHandler)

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
