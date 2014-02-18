package library

import (
	"errors"
	"github.com/nytlabs/streamtools/st/blocks" // blocks
	"os"
)

// specify those channels we're going to use to communicate with streamtools
type ToFile struct {
	blocks.Block
	file      *os.File
	queryrule chan chan interface{}
	inrule    chan interface{}
	inpoll    chan interface{}
	in        chan interface{}
	out       chan interface{}
	quit      chan interface{}
}

// we need to build a simple factory so that streamtools can make new blocks of this kind
func NewToFile() blocks.BlockInterface {
	return &ToFile{}
}

// Setup is called once before running the block. We build up the channels and specify what kind of block this is.
func (b *ToFile) Setup() {
	b.Kind = "ToFile"
	b.in = b.InRoute("in")
	b.inrule = b.InRoute("rule")
	b.queryrule = b.QueryRoute("rule")
	b.quit = b.InRoute("quit")
	b.out = b.Broadcast()
}

// Run is the block's main loop. Here we listen on the different channels we set up.
func (b *ToFile) Run() {
	for {
		select {
		case msgI := <-b.inrule:
			// convert message to map[string]string
			msg := msgI.(map[string]string)
			// set a parameter of the block
			filename, ok := msg["Filename"]
			if !ok {
				b.Error(errors.New("Rule message did not contain Filename"))
			}
			fo, err := os.Create(filename)
			if err != nil {
				b.Error(err)
			}
			// set the new file
			b.file = fo
		case <-b.quit:
			// quit the block
			return
		case <-b.in:
			// deal with inbound data
		case <-b.queryrule:
			// deal with a query request
		}
	}
}
