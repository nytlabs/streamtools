package library

import (
	"github.com/nytlabs/streamtools/st/blocks" // blocks
	"github.com/nytlabs/streamtools/st/util"
)

// specify those channels we're going to use to communicate with streamtools
type PackByCount struct {
	blocks.Block
	queryrule chan blocks.MsgChan
	inrule    blocks.MsgChan
	clear     blocks.MsgChan
	flush     blocks.MsgChan
	in        blocks.MsgChan
	out       blocks.MsgChan
	quit      blocks.MsgChan
}

// we need to build a simple factory so that streamtools can make new blocks of this kind
func NewPackByCount() blocks.BlockInterface {
	return &PackByCount{}
}

// Setup is called once before running the block. We build up the channels and specify what kind of block this is.
func (b *PackByCount) Setup() {
	b.Kind = "Core"
	b.Desc = "Packs incoming messages into array. When the array is filled, it is emitted."
	b.in = b.InRoute("in")
	b.inrule = b.InRoute("rule")
	b.queryrule = b.QueryRoute("rule")
	b.clear = b.InRoute("clear")
	b.flush = b.InRoute("flush")
	b.quit = b.Quit()
	b.out = b.Broadcast()
}

// Run is the block's main loop. Here we listen on the different channels we set up.
func (b *PackByCount) Run() {
	var batch []interface{}
	var packSize int

	for {
		select {
		case ruleI := <-b.inrule:
			packSizeTmp, err := util.ParseFloat(ruleI, "MaxCount")
			if err != nil {
				b.Error("error parsing batch size")
				break
			}

			packSize = int(packSizeTmp)
			batch = nil

		case <-b.quit:
			// quit the block
			return
		case m := <-b.in:
			if packSize == 0 {
				break
			}

			if len(batch) == packSize {
				b.out <- map[string]interface{}{
					"Pack": batch,
				}
				batch = nil
			}

			batch = append(batch, m)

		case <-b.clear:
			batch = nil
		case <-b.flush:
			b.out <- map[string]interface{}{
				"Pack": batch,
			}
			batch = nil
		case r := <-b.queryrule:
			r <- map[string]interface{}{
				"MaxCount": packSize,
			}
		}
	}
}
