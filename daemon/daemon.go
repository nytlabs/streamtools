package daemon

import (
	"fmt"
	"github.com/bitly/go-simplejson"
	"github.com/nytlabs/streamtools/blocks"
	"log"
	"net/http"
	"strings"
)

var (
	// channel that returns the next ID
	idChan chan string
)

// hub keeps track of all the blocks and connections
type Daemon struct {
	blockMap      map[string]*blocks.Block
}

// The rootHandler returns information about the whole system
func (d *Daemon) rootHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "hello! this is streamtools")
	fmt.Fprintln(w, "ID: BlockType")
	for id, block := range d.blockMap {
		fmt.Fprintln(w, id+":", block.Template.BlockType)
	}
}

// The createHandler creates new blocks
func (d *Daemon) createHandler(w http.ResponseWriter, r *http.Request) {
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
		d.CreateBlock(blockType[0], id)

	} else {
		log.Println("no blocktype specified")
	}
}

// The connectHandler connects together two blocks
func (d *Daemon) connectHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Println("could not parse form on /connect")
	}
	from := r.Form["from"][0]
	to := r.Form["to"][0]
	log.Println("connecting", from, "to", to)
	d.CreateConnection(from, to)
}

// The routeHandler deals with any incoming message sent to an arbitrary block endpoint
func (d *Daemon) routeHandler(w http.ResponseWriter, r *http.Request) {
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
	ResponseChan := make(chan *simplejson.Json)
	blockRouteChan := d.blockMap[id].Routes[route]
	blockRouteChan <- blocks.RouteResponse{
		Msg:          msg,
		ResponseChan: ResponseChan,
	}
	blockMsg := <-ResponseChan
	out, err := blockMsg.MarshalJSON()
	if err != nil {
		log.Println(err.Error())
	}

	fmt.Fprint(w, string(out))
}

func (d *Daemon) libraryHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "libraryBlob")
}

func (d *Daemon) createRoutes(b *blocks.Block){
	for _, routeName := range b.Template.RouteNames {
		log.Println("creating route /blocks/"+b.ID+"/"+routeName)
		http.HandleFunc("/blocks/"+b.ID+"/"+routeName, d.routeHandler)
	}
}

func (d *Daemon) CreateConnection(from string, to string) {

	ID := <-idChan
	conn := blocks.NewBlock("connection", ID)

	// channel from the inbound block to the connection
	fromBlock := d.blockMap[from]
	inChan := make(chan *simplejson.Json)
	fromBlock.OutChans[ID] = inChan
	conn.InChan = inChan

	// channel from the connection to the outbound block
	toBlock := d.blockMap[to]
	inChan = toBlock.InChan
	conn.OutChans[toBlock.ID] = inChan

	d.createRoutes(conn)
	d.blockMap[conn.ID] = conn
	go blocks.Library["connection"].Routine(conn)
}

func (d *Daemon) CreateBlock(name string, ID string) {
	b := blocks.NewBlock(name, ID)
	d.createRoutes(b)
	d.blockMap[b.ID] = b
	go blocks.Library[name].Routine(b)
}

func (d *Daemon) Run(port string) {

	// start the ID Service
	idChan = make(chan string)
	go IDService(idChan)

	// start the library service
	blocks.BuildLibrary()

	// initialise the block maps
	d.blockMap = make(map[string]*blocks.Block)

	// instantiate the base handlers
	http.HandleFunc("/", d.rootHandler)
	http.HandleFunc("/create", d.createHandler)
	http.HandleFunc("/connect", d.connectHandler)
	http.HandleFunc("/library", d.libraryHandler)

	// start the http server
	log.Println("starting stream tools on port",  port)
	err := http.ListenAndServe(":"+ port, nil)
	if err != nil {
		log.Fatalf(err.Error())
	}
}
