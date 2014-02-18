package library

import (
	"github.com/nytlabs/streamtools/st/blocks" // blocks
	"testing"
)

func newBlock(id, kind string) blocks.BlockInterface {

	library := map[string]func() blocks.BlockInterface{
		"count": NewCount,
	}

	chans := blocks.BlockChans{
		InChan:    make(chan *blocks.Msg),
		QueryChan: make(chan *blocks.QueryMsg),
		AddChan:   make(chan *blocks.AddChanMsg),
		DelChan:   make(chan *blocks.Msg),
		ErrChan:   make(chan error),
	}

	// actual block
	b := library[kind]()
	b.Build(chans)

	return b

}

func TestCount(t *testing.T) {
	b := newBlock("testingCount", "count")
	blocks.BlockRoutine(b)
}
