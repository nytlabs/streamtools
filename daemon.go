package streamtools

import (
	"flag"
	"fmt"
	"log"
	"net/http"
)

var (
	h      hub
	idChan chan string
	port   = flag.String("port", "7070", "stream tools port")
)

type StreamtoolsQuery struct {
	w http.ResponseWriter
	r *http.Request
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "hello! This is streamtools.")
}

func createHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Println("could not parse form on /create")
	}
	if blockType, ok := r.Form["blockType"]; ok {
		h.CreateBlock(blockType[0])
	} else {
		log.Println("no blockType specified")
	}
}

func connectHandler(w http.ResponseWriter, r *http.Request) {

}

func Run() {
	idChan = make(chan string)
	go IDService(idChan)
	buildLibrary()
	h = hub{
		connectionMap: make(map[string]*Connection),
		blockMap:      make(map[string]*Block),
	}
	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/create", createHandler)
	http.HandleFunc("/connect", connectHandler)
	log.Println("starting streamtools server on port", port)
	err := http.ListenAndServe(":"+*port, nil)
	if err != nil {
		log.Println(err)
	}
}
