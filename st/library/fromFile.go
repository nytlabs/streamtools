package library

import (
	"bufio"
	"encoding/json"
	"io"
	"os"

	"github.com/nytlabs/streamtools/st/blocks" // blocks
	"github.com/nytlabs/streamtools/st/util"
)

// specify those channels we're going to use to communicate with streamtools
type FromFile struct {
	blocks.Block
	queryrule chan blocks.MsgChan
	inrule    blocks.MsgChan
	in        blocks.MsgChan
	inpoll    blocks.MsgChan
	out       blocks.MsgChan
	quit      blocks.MsgChan
}

// we need to build a simple factory so that streamtools can make new blocks of this kind
func NewFromFile() blocks.BlockInterface {
	return &FromFile{}
}

// Setup is called once before running the block. We build up the channels and specify what kind of block this is.
func (b *FromFile) Setup() {
	b.Kind = "Data Stores"
	b.Desc = "reads in a file specified by the block's rule, emitting a message for each line"
	b.in = b.InRoute("in")
	b.inrule = b.InRoute("rule")
	b.inpoll = b.InRoute("poll")
	b.queryrule = b.QueryRoute("rule")
	b.quit = b.Quit()
	b.out = b.Broadcast()
}

// Run is the block's main loop. Here we listen on the different channels we set up.
func (b *FromFile) Run() {
	var file *os.File
	var filename string
	var err error
	var reader *bufio.Reader

	for {
		select {
		case msgI := <-b.inrule:
			// set a parameter of the block
			filename, err = util.ParseString(msgI, "Filename")
			if err != nil {
				b.Error(err)
				continue
			}

			file, err = os.Open(filename)
			if err != nil {
				b.Error(err)
				continue
			}

			reader = bufio.NewReader(file)

		case c := <-b.queryrule:
			c <- map[string]interface{}{
				"Filename": filename,
			}

		case <-b.inpoll:
			if reader == nil && filename == "" {
				b.Error("you must configure a filename before polling this block.")
				break
			}

			var outMsg interface{}

			line, err := reader.ReadBytes('\n')
			if err != nil && err != io.EOF {
				b.Error(err)
				continue
			}

			err = json.Unmarshal(line, &outMsg)
			// if the json parsing fails, store data unparsed as "data"
			if err != nil {
				outMsg = map[string]interface{}{
					"data": string(line),
				}
			}

			b.out <- outMsg

		case <-b.quit:
			// quit the block
			if file != nil {
				file.Close()
			}
			return
		}
	}
}
