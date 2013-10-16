package streamtools

import (
	"flag"
	"fmt"
	"log"
	"net/http"
)

var (
	idChan chan string
	port   = flag.String("port", "7070", "stream tools port")
)

type StreamtoolsQuery struct {
	w http.ResponseWriter
	r *http.Request
}

type hub struct {
	connectionMap map[string]Connection
	blockMap      map[string]Block
}

func (self *hub) rootHandler(w http.ResponseWriter, r *http.Request) {

}

func (self *hub) createHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Println("could not parse form on /create")
	}

	if blockType, ok := r.Form["blockType"]; ok {
		self.CreateBlock(blockType[0])
	}
}

func (self *hub) connectHandler(w http.ResponseWriter, r *http.Request) {

}

func (self *hub) CreateBlock(blockType string) {
	block := NewBlock(blockType)
	self.blockMap[blockType+"_"+block.getID()] = block
	go block.blockRoutine()
}

func (self *hub) Run() {
	idChan = make(chan string)
	go IDService(idChan)
	buildLibrary()

	self.connectionMap = make(map[string]Connection)
	self.blockMap = make(map[string]Block)

	http.HandleFunc("/", self.rootHandler)
	http.HandleFunc("/create", self.createHandler)
	http.HandleFunc("/connect", self.connectHandler)

	log.Println("staring stramtools on ", *port)
	err := http.ListenAndServe(":"+*port, nil)

	if err != nil {
		log.Println(err)
	}
}

func Run() {
	h := hub{}
	h.Run()
}
