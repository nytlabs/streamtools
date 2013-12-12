package daemon

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ant0ine/go-json-rest"
	"github.com/nytlabs/streamtools/blocks"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
)

const (
	READ_MAX = 1024768
)

var (
	// channel that returns the next ID
	idChan chan string
)

// Daemon keeps track of all the blocks and connections
type Daemon struct {
	blockMap map[string]*blocks.Block
}

// The rootHandler returns information about the whole system
func (d *Daemon) rootHandler(w *rest.ResponseWriter, r *rest.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Write(index())
}

// The createHandler creates new blocks
func (d *Daemon) createHandler(w *rest.ResponseWriter, r *rest.Request) {
	err := r.ParseForm()
	if err != nil {
		ApiResponse(w, 500, "BAD_REQUEST")
		return
	}

	var id string
	var blockType string
	fType, typeExists := r.Form["blockType"]
	fID, idExists := r.Form["id"]

	if typeExists == false {
		ApiResponse(w, 500, "MISSING_BLOCKTYPE")
		return
	} else {
		blockType = fType[0]
	}

	_, inLibrary := blocks.Library[blockType]
	if inLibrary == false {
		ApiResponse(w, 500, "INVALID_BLOCKTYPE")
		return
	}

	if idExists == false {
		id = <-idChan
	} else {

		if len(strings.TrimSpace(fID[0])) == 0 {
			ApiResponse(w, 500, "BAD_BLOCK_ID")
			return
		}

		_, notUnique := d.blockMap[fID[0]]
		if notUnique == true {
			ApiResponse(w, 500, "BLOCK_ID_ALREADY_EXISTS")
			return
		} else {
			id = fID[0]
		}
	}

	d.CreateBlock(blockType, id)

	ApiResponse(w, 200, "BLOCK_CREATED")
}

func (d *Daemon) deleteHandler(w *rest.ResponseWriter, r *rest.Request) {
	err := r.ParseForm()
	if err != nil {
		ApiResponse(w, 500, "BAD_REQUEST")
		return
	}

	_, hasID := r.Form["id"]
	if hasID == false {
		ApiResponse(w, 500, "MISSING_BLOCK_ID")
		return
	}

	id := r.Form["id"][0]

	err = d.DeleteBlock(id)

	if err != nil {
		ApiResponse(w, 500, err.Error())
		return
	}

	ApiResponse(w, 200, "BLOCK_DELETED")
}

func (d *Daemon) DeleteBlock(id string) error {
	block, ok := d.blockMap[id]
	if ok == false {
		return errors.New("BLOCK_NOT_FOUND")
	}

	// delete inbound channels
	for k, _ := range block.InBlocks {
		d.blockMap[k].AddChan <- &blocks.OutChanMsg{
			Action: blocks.DELETE_OUT_CHAN,
			ID:     block.ID,
		}

		log.Println("disconnecting \"" + block.ID + "\" from \"" + k + "\"")

		delete(d.blockMap[k].OutBlocks, block.ID)

		if d.blockMap[k].BlockType == "connection" {
			d.DeleteBlock(k)
		}
	}

	// delete outbound channels
	for k, _ := range block.OutBlocks {
		delete(d.blockMap[k].InBlocks, block.ID)
		if d.blockMap[k].BlockType == "connection" {
			d.DeleteBlock(k)
		}
	}

	// delete the block itself
	block.QuitChan <- true
	delete(d.blockMap, id)

	return nil
}

// The connectHandler connects together two blocks
func (d *Daemon) connectHandler(w *rest.ResponseWriter, r *rest.Request) {
	var id string

	err := r.ParseForm()
	if err != nil {
		ApiResponse(w, 500, "BAD_REQUEST")
		return
	}

	_, hasFrom := r.Form["from"]
	if hasFrom == false {
		ApiResponse(w, 500, "MISSING_FROM_BLOCK_ID")
		return
	}

	_, hasTo := r.Form["to"]
	if hasTo == false {
		ApiResponse(w, 500, "MISSING_TO_BLOCK_ID")
		return
	}

	fID, hasID := r.Form["id"]
	if hasID == false {
		id = <-idChan
	} else {

		if len(strings.TrimSpace(fID[0])) == 0 {
			ApiResponse(w, 500, "BAD_CONNECTION_ID")
			return
		}

		_, ok := d.blockMap[fID[0]]
		if ok == false {
			id = fID[0]
		} else {
			ApiResponse(w, 500, "BLOCK_ID_ALREADY_EXISTS")
			return
		}
	}

	from := r.Form["from"][0]
	to := r.Form["to"][0]

	if len(from) == 0 {
		ApiResponse(w, 500, "MISSING_FROM_BLOCK_ID")
		return
	}

	if len(to) == 0 {
		ApiResponse(w, 500, "MISSING_TO_BLOCK_ID")
		return
	}

	_, exists := d.blockMap[from]
	if exists == false {
		ApiResponse(w, 500, "FROM_BLOCK_NOT_FOUND")
		return
	}

	_, exists = d.blockMap[strings.Split(to, "/")[0]]
	if exists == false {
		ApiResponse(w, 500, "TO_BLOCK_NOT_FOUND")
		return
	}

	err = d.CreateConnection(from, to, id)
	if err != nil {
		ApiResponse(w, 500, "TO_ROUTE_NOT_FOUND")
		return
	}

	ApiResponse(w, 200, "CONNECTION_CREATED")
}

// The routeHandler deals with any incoming message sent to an arbitrary block endpoint
func (d *Daemon) routeHandler(w *rest.ResponseWriter, r *rest.Request) {
	//id := strings.Split(r.URL.Path, "/")[2]
	//route := strings.Split(r.URL.Path, "/")[3]

	id, ok := r.PathParams["id"]
	if ok == false {
		ApiResponse(w, 500, "MISSING_BLOCK_ID")
		return
	}

	route, ok := r.PathParams["route"]
	if ok == false {
		ApiResponse(w, 500, "MISSING_ROUTE")
		return
	}

	_, ok = d.blockMap[id]
	if ok == false {
		ApiResponse(w, 500, "BLOCK_ID_NOT_FOUND")
		return
	}

	_, ok = d.blockMap[id].Routes[route]
	if ok == false {
		ApiResponse(w, 500, "ROUTE_NOT_FOUND")
		return
	}

	msg, err := ioutil.ReadAll(io.LimitReader(r.Body, READ_MAX))

	if err != nil {
		ApiResponse(w, 500, "BAD_REQUEST")
		return
	}

	var outMsg blocks.BMsg

	if len(msg) > 0 {
		err = json.Unmarshal(msg, &outMsg)
		if err != nil {
			log.Println(msg)
			ApiResponse(w, 500, "BAD_JSON")
			return
		}
	}

	ResponseChan := make(chan blocks.BMsg)
	blockRouteChan := d.blockMap[id].Routes[route]
	blockRouteChan <- blocks.RouteResponse{
		Msg:          outMsg,
		ResponseChan: ResponseChan,
	}
	respMsg := <-ResponseChan

	respJson, err := json.Marshal(respMsg)
	if err != nil {
		ApiResponse(w, 500, "BAD_RESPONSE_FROM_BLOCK")
		return
	}

	DataResponse(w, respJson)
}

func (d *Daemon) libraryHandler(w *rest.ResponseWriter, r *rest.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Content-Length", strconv.Itoa(len(blocks.LibraryBlob)))
	fmt.Fprint(w, blocks.LibraryBlob)
}

func (d *Daemon) listHandler(w *rest.ResponseWriter, r *rest.Request) {
	blockList := []map[string]interface{}{}
	for _, v := range d.blockMap {
		blockItem := make(map[string]interface{})
		blockItem["BlockType"] = v.BlockType
		blockItem["ID"] = v.ID
		blockItem["InBlocks"] = []string{}
		blockItem["OutBlocks"] = []string{}
		blockItem["Routes"] = []string{}
		for k, _ := range v.InBlocks {
			blockItem["InBlocks"] = append(blockItem["InBlocks"].([]string), k)
		}
		for k, _ := range v.OutBlocks {
			blockItem["OutBlocks"] = append(blockItem["OutBlocks"].([]string), k)
		}
		for k, _ := range v.Routes {
			blockItem["Routes"] = append(blockItem["Routes"].([]string), k)
		}
		blockList = append(blockList, blockItem)
	}
	blob, _ := json.Marshal(blockList)

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Content-Length", strconv.Itoa(len(blob)))
	fmt.Fprint(w, string(blob))
}

func (d *Daemon) CreateConnection(from string, to string, ID string) error {
	d.CreateBlock("connection", ID)

	d.blockMap[from].AddChan <- &blocks.OutChanMsg{
		Action:  blocks.CREATE_OUT_CHAN,
		OutChan: d.blockMap[ID].InChan,
		ID:      ID,
	}

	/*d.blockMap[ID].AddChan <- &blocks.OutChanMsg{
		Action:  blocks.CREATE_OUT_CHAN,
		OutChan: d.blockMap[to].InChan,
		ID:      to,
	}*/
	toParts := strings.Split(to, "/")
	switch len(toParts) {
	case 1:
		d.blockMap[ID].AddChan <- &blocks.OutChanMsg{
			Action:  blocks.CREATE_OUT_CHAN,
			OutChan: d.blockMap[to].InChan,
			ID:      to,
		}
	case 2:
		d.blockMap[ID].AddChan <- &blocks.OutChanMsg{
			Action:  blocks.CREATE_OUT_CHAN,
			OutChan: d.blockMap[toParts[0]].Routes[toParts[1]],
			ID:      to,
		}
	default:
		err := errors.New("malformed to route specification")
		return err
	}

	// add the from block to the list of inblocks for connection.
	d.blockMap[from].OutBlocks[ID] = true
	d.blockMap[ID].InBlocks[from] = true
	d.blockMap[ID].OutBlocks[toParts[0]] = true
	d.blockMap[toParts[0]].InBlocks[ID] = true

	log.Println("connected", d.blockMap[from].ID, "to", d.blockMap[toParts[0]].ID)

	return nil
}

func (d *Daemon) CreateBlock(name string, ID string) {
	// TODO: Clean this up.
	//
	// In order to avoid data races the blocks held in daemon's blockMap
	// are not the same blocks held in each block routine. When CreateBlock
	// is called, we actually create two blocks: one to store in daemon's
	// blockMap and one to send to the block routine.
	//
	// The block stored in daemon's blockmap doesn't make use of OutChans as
	// a block's OutChans can be dynamically modified when connections are
	// added or deleted. All of the other fields, such as ID, name, and all
	// the channels that go into the block (inChan, Routes) are the SAME
	// in both the daemon blockMap block and the blockroutine block.
	//
	// Becauase of this very minor difference it would be a huge semantic help
	// if the type going to the blockroutines was actually different than the
	// type being kept in daemon's blockmap.
	//
	// Modifications to blocks in daemon's blockMap will obviously not
	// proliferate to blockroutines and all changes (such as adding outchans)
	// can only be done through messages. A future daemon block type might
	// want to restrict how daemon blocks can be used, such as creating
	// getters and no setters. Or perhaps a setter automatically takes care
	// of sending a message to the blockroutine to emulate the manipulation
	// of a single variable.

	// create the block that will be stored in blockMap
	b, _ := blocks.NewBlock(name, ID)
	//d.createRoutes(b)
	d.blockMap[b.ID] = b
	d.blockMap[b.ID].InBlocks = make(map[string]bool)
	d.blockMap[b.ID].OutBlocks = make(map[string]bool)

	// create the block that will be sent to the blockroutine and copy all
	// chan references from the previously created block
	c, _ := blocks.NewBlock(name, ID)
	for k, v := range b.Routes {
		c.Routes[k] = v
	}

	c.InChan = b.InChan
	c.AddChan = b.AddChan
	c.QuitChan = b.QuitChan

	//create outchans for use only by blockroutine block.
	c.OutChans = make(map[string]chan blocks.BMsg)

	go blocks.Library[name].Routine(c)

	log.Println("started block \"" + ID + "\" of type " + name)
}

func (d *Daemon) Run(port string) {

	// start the ID Service
	idChan = make(chan string)
	go IDService(idChan)

	// start the library service
	blocks.BuildLibrary()

	// initialise the block maps
	d.blockMap = make(map[string]*blocks.Block)

	handler := rest.ResourceHandler{
		EnableRelaxedContentType: true,
	}

	//TODO: make this a _real_ restful API
	handler.SetRoutes(
		rest.Route{"GET", "/", d.rootHandler},
		rest.Route{"GET", "/library", d.libraryHandler},
		rest.Route{"GET", "/list", d.listHandler},
		rest.Route{"GET", "/create", d.createHandler},
		rest.Route{"GET", "/delete", d.deleteHandler},
		rest.Route{"GET", "/connect", d.connectHandler},
		rest.Route{"GET", "/blocks/:id/:route", d.routeHandler},
		rest.Route{"POST", "/blocks/:id/:route", d.routeHandler},
	)
	log.Println("starting stream tools on port", port)
	err := http.ListenAndServe(":"+port, &handler)
	if err != nil {
		log.Fatalf(err.Error())
	}
}
