package library

import (
	"bufio"
	"encoding/json"
	"github.com/nytlabs/streamtools/st/blocks" // blocks
	"github.com/nytlabs/streamtools/st/util"
	"os"
)

// specify those channels we're going to use to communicate with streamtools
type FromFile struct {
	blocks.Block
	queryrule chan chan interface{}
	inrule    chan interface{}
	in        chan interface{}
	out       chan interface{}
	quit      chan interface{}
}

// we need to build a simple factory so that streamtools can make new blocks of this kind
func NewFromFile() blocks.BlockInterface {
	return &FromFile{}
}

// Setup is called once before running the block. We build up the channels and specify what kind of block this is.
func (b *FromFile) Setup() {
	b.Kind = "FromFile"
	b.Desc = "reads in a file specified by the block's rule, emitting a message for each line"
	b.in = b.InRoute("in")
	b.inrule = b.InRoute("rule")
	b.queryrule = b.QueryRoute("rule")
	b.quit = b.Quit()
	b.out = b.Broadcast()
}

// Run is the block's main loop. Here we listen on the different channels we set up.
func (b *FromFile) Run() {
	var file *os.File
	var filename string

	for {
		select {
		case msgI := <-b.inrule:
			// set a parameter of the block
			filename, err := util.ParseString(msgI, "Filename")
			if err != nil {
				b.Error(err)
				continue
			}

			file, err := os.Open(filename)
			if err != nil {
				b.Error(err)
				continue
			}

			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				var outMsg interface{}
				lineBytes := scanner.Bytes()
				err := json.Unmarshal(lineBytes, &outMsg)
				// if the json parsing fails, store data unparsed as "data"
				if err != nil {
					outMsg = map[string]interface{}{
						"data": lineBytes,
					}
				}
				b.out <- outMsg
			}

			if err := scanner.Err(); err != nil {
				b.Error(err)
				continue
			}
		case <-b.quit:
			// quit the block
			if file != nil {
				file.Close()
			}
			return

		case c := <-b.queryrule:
			c <- map[string]interface{}{
				"Filename": filename,
			}
		}
	}
}
