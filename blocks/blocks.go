package blocks

import (
	"errors"
	"github.com/bitly/go-simplejson"
)

const (
	CREATE_OUT_CHAN = iota
	DELETE_OUT_CHAN = iota
)

// Block is the basic interface for processing units in streamtools
type BlockTemplate struct {
	BlockType  string
	RouteNames []string
	// BlockRoutine is the central processing routine for a block. All the work gets done in here
	Routine BlockRoutine
}

type Block struct {
	BlockType string
	ID        string
	InChan    chan *simplejson.Json
	OutChans  map[string]chan *simplejson.Json
	Routes    map[string]chan RouteResponse
	AddChan   chan *OutChanMsg
	InBlocks  map[string]bool // bool is dumb.
	OutBlocks map[string]bool // bool is dumb.
	QuitChan  chan bool
}

type OutChanMsg struct {
	// type of action to perform
	Action int
	// new channel to introduce to a block's outChan array
	OutChan chan *simplejson.Json
	// ID of the connection block
	ID string
}

type BlockRoutine func(*Block)

// RouteResponse is passed into a block to query via established handlers
type RouteResponse struct {
	Msg          []byte
	ResponseChan chan []byte
}

func NewBlock(name string, ID string) (*Block, error) {
	routes := make(map[string]chan RouteResponse)

	if _, ok := Library[name]; !ok {
		return nil, errors.New("cannot find " + name + " in the Library")
	}

	for _, name := range Library[name].RouteNames {
		routes[name] = make(chan RouteResponse)
	}

	b := &Block{
		BlockType: name,
		ID:        ID,
		InChan:    make(chan *simplejson.Json),
		Routes:    routes,
		AddChan:   make(chan *OutChanMsg),
		QuitChan:  make(chan bool),
	}

	return b, nil
}
