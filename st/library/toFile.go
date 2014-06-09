package library

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"

	"github.com/nytlabs/streamtools/st/blocks" // blocks
	"github.com/nytlabs/streamtools/st/util"
)

// specify those channels we're going to use to communicate with streamtools
type ToFile struct {
	blocks.Block
	queryrule chan blocks.MsgChan
	inrule    blocks.MsgChan
	in        blocks.MsgChan
	out       blocks.MsgChan
	quit      blocks.MsgChan
}

// we need to build a simple factory so that streamtools can make new blocks of this kind
func NewToFile() blocks.BlockInterface {
	return &ToFile{}
}

// Setup is called once before running the block. We build up the channels and specify what kind of block this is.
func (b *ToFile) Setup() {
	b.Kind = "Data Stores"
	b.Desc = "writes messages, separated by newlines, to a file on the local filesystem"
	b.in = b.InRoute("in")
	b.inrule = b.InRoute("rule")
	b.queryrule = b.QueryRoute("rule")
	b.quit = b.Quit()
	b.out = b.Broadcast()
}

// Run is the block's main loop. Here we listen on the different channels we set up.
func (b *ToFile) Run() {
	var err error
	var file *os.File
	var filename string

	for {
		select {
		case msgI := <-b.inrule:
			filename, err = util.ParseString(msgI, "Filename")
			if err != nil {
				b.Error(err)
			}

			file, err = os.Create(filename)
			if err != nil {
				b.Error(err)
			}

		case <-b.quit:
			// quit the block
			if file != nil {
				file.Close()
			}
			return
		case msg := <-b.in:
			// deal with inbound data
			writer := bufio.NewWriter(file)
			msgStr, err := json.Marshal(msg)
			if err != nil {
				b.Error(err)
				continue
			}
			fmt.Fprintln(writer, string(msgStr))
			writer.Flush()

		case MsgChan := <-b.queryrule:
			// deal with a query request
			MsgChan <- map[string]interface{}{
				"Filename": filename,
			}
		}
	}
}
