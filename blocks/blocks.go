package blocks

import (
	"errors"
)

const (
	CREATE_OUT_CHAN = iota
	DELETE_OUT_CHAN = iota
)

type BMsg interface{}

func Set(m interface{}, key string, val interface{}) error {
	min, ok := m.(map[string]interface{})
	if !ok {
		return errors.New("type assertion failed")
	}
	min[key] = val
	return nil
}

func Get(msg interface{}, branch ...string) (interface{}, error) {
	min := msg
	for i := range branch {
		m, ok := min.(map[string]interface{})
		if !ok {
			return nil, errors.New("type assertion failed")
		}
		if val, ok := m[branch[i]]; ok {
			min = val
		} else {
			return nil, errors.New("cannot find branch")
		}
	}
	return min, nil
}

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
	InChan    chan BMsg
	OutChans  map[string]chan BMsg
	Routes    map[string]chan BMsg
	AddChan   chan *OutChanMsg
	InBlocks  map[string]bool // bool is dumb.
	OutBlocks map[string]bool // bool is dumb.
	QuitChan  chan bool
}

type OutChanMsg struct {
	// type of action to perform
	Action int
	// new channel to introduce to a block's outChan array
	OutChan chan BMsg
	// ID of the connection block
	ID string
}

type BlockRoutine func(*Block)

// RouteResponse is passed into a block to query via established handlers
type RouteResponse struct {
	Msg          BMsg
	ResponseChan chan BMsg 
}

func NewBlock(name string, ID string) (*Block, error) {
	routes := make(map[string]chan BMsg)

	if _, ok := Library[name]; !ok {
		return nil, errors.New("cannot find " + name + " in the Library")
	}

	for _, name := range Library[name].RouteNames {
		routes[name] = make(chan BMsg)
	}

	b := &Block{
		BlockType: name,
		ID:        ID,
		InChan:    make(chan BMsg),
		Routes:    routes,
		AddChan:   make(chan *OutChanMsg),
		QuitChan:  make(chan bool),
	}

	return b, nil
}
