package blocks

import (
	"github.com/bitly/go-simplejson"
)

// Block is the basic interface for processing units in streamtools
type Block struct {
	Template *BlockTemplate
	// BlockRoutine is the central processing routine for a block. All the work gets done in here
	Routine BlockRoutine
}

type BlockTemplate struct {
	BlockType  string
	RouteNames []string
}

type BlockDefinition struct {
	Template BlockTemplate
	ID       string
	InChan   chan *simplejson.Json
	OutChans map[string]chan *simplejson.Json
	Routes   map[string]chan RouteResponse
}

type BlockRoutine func(*BlockDefinition)

// RouteResponse is passed into a block to query via established handlers
type RouteResponse struct {
	Msg          *simplejson.Json
	ResponseChan chan *simplejson.Json
}

func NewBlock(name string, ID string) *BlockDefinition {

	d := Library[name].Template

	routes := make(map[string]chan RouteResponse)
	for _, name := range d.RouteNames {
		routes[name] = make(chan RouteResponse)
	}

	b := &BlockDefinition{
		Template: *d,
		ID:       ID,
		InChan:   make(chan *simplejson.Json),
		OutChans: make(map[string]chan *simplejson.Json),
		Routes:   routes,
	}

	return b
}
