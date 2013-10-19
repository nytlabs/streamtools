package blocks

import (
	"github.com/bitly/go-simplejson"
)

// Block is the basic interface for processing units in streamtools
type BlockTemplate struct {
	BlockType  string
	RouteNames []string
	// BlockRoutine is the central processing routine for a block. All the work gets done in here
	Routine BlockRoutine
}

type Block struct {
	Template *BlockTemplate
	ID       string
	InChan   chan *simplejson.Json
	OutChans map[string]chan *simplejson.Json
	Routes   map[string]chan RouteResponse
}

type BlockRoutine func(*Block)

// RouteResponse is passed into a block to query via established handlers
type RouteResponse struct {
	Msg          *simplejson.Json
	ResponseChan chan *simplejson.Json
}

func NewBlock(name string, ID string) *Block {
	routes := make(map[string]chan RouteResponse)

	for _, name := range Library[name].RouteNames {
		routes[name] = make(chan RouteResponse)
	}

	b := &Block{
		Template: Library[name],
		ID:       ID,
		InChan:   make(chan *simplejson.Json),
		OutChans: make(map[string]chan *simplejson.Json),
		Routes:   routes,
	}

	return b
}
