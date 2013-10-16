package streamtools

import (
    "net/http"
    "log"
)

var (
    h hub
    idChan chan string
)

func rootHandler(w http.ResponseWriter, r *http.Request){

}

func createHandler(w http.ResponseWriter, r *http.Request){
    err := r.ParseForm()

    if err != nil {
        log.Println("could not parse form on /create")
    }

    if blockType, ok := r.Form["blockType"]; ok{
        h.CreateBlock(blockType[0])
    }

}

func connectHandler(w http.ResponseWriter, r *http.Request){

}

func Run(){
    idChan = make(chan string)
    go IDService(idChan)
    buildLibrary()

    h = hub {
        connectionMap: make(map[string]*Connection),
        blockMap     : make(map[string]*Block),
    }

    log.Println("run")   
    http.HandleFunc("/", rootHandler)
    http.HandleFunc("/create", createHandler)
    http.HandleFunc("/connect", connectHandler)

    err := http.ListenAndServe(":7070", nil)

    if err != nil {
        log.Println(err)
    }
}