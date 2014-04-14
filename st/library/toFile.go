package library

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/nytlabs/streamtools/st/blocks" // blocks
	"github.com/nytlabs/streamtools/st/util"
	"os"
)

// specify those channels we're going to use to communicate with streamtools
type ToFile struct {
	blocks.Block
	file      *os.File
	filename  string
	queryrule chan chan interface{}
	inrule    chan interface{}
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
	b.Desc = "writes messages, separated by newlines, to a file on the local filesystem"
	b.in = b.InRoute("in")
	b.inrule = b.InRoute("rule")
	b.queryrule = b.QueryRoute("rule")
	b.quit = b.Quit()
	b.out = b.Broadcast()
}

// Run is the block's main loop. Here we listen on the different channels we set up.
func (b *ToFile) Run() {
	for {
		select {
		case msgI := <-b.inrule:
			// set a parameter of the block
			filename, _ := util.ParseString(msgI, "Filename")

			fo, err := os.Create(filename)
			if err != nil {
				b.Error(err)
			}
			// set the new file
			b.file = fo
			// record the filename
			b.filename = filename
		case <-b.quit:
			// quit the block
			if b.file != nil {
				b.file.Close()
			}
			return
		case msg := <-b.in:
			// deal with inbound data
			w := bufio.NewWriter(b.file)
			msgStr, err := json.Marshal(msg)
			if err != nil {
				b.Error(err)
				continue
			}
			fmt.Fprintln(w, string(msgStr))
			w.Flush()
		case respChan := <-b.queryrule:
			// deal with a query request
			respChan <- map[string]interface{}{
				"Filename": b.filename,
			}
		}
	}
}
