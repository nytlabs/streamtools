// +build arm

package library

import (
	"github.com/mrmorphic/hwio"                // hwio
	"github.com/nytlabs/streamtools/st/blocks" // blocks
	"github.com/nytlabs/streamtools/st/util"
)

// specify those channels we're going to use to communicate with streamtools
type AnalogPin struct {
	blocks.Block
	queryrule chan blocks.MsgChan
	inrule    blocks.MsgChan
	inpoll    blocks.MsgChan
	out       blocks.MsgChan
	quit      blocks.MsgChan
}

// we need to build a simple factory so that streamtools can make new blocks of this kind
func NewAnalogPin() blocks.BlockInterface {
	return &AnalogPin{}
}

// Setup is called once before running the block. We build up the channels and specify what kind of block this is.
func (b *AnalogPin) Setup() {
	b.Kind = "Hardware I/O"
	b.Desc = "(embedded applications) returns current state of the pin"
	b.inrule = b.InRoute("rule")
	b.inpoll = b.InRoute("poll")
	b.queryrule = b.QueryRoute("rule")
	b.quit = b.Quit()
	b.out = b.Broadcast()
}

// Run is the block's main loop. Here we listen on the different channels we set up.
func (b *AnalogPin) Run() {
	var pin hwio.Pin
	var pinStr string
	var err error
	// Get the module
	m, e := hwio.GetAnalogModule()
	if e != nil {
		b.Log(e)
	}
	// Enable it.
	e = m.Enable()
	if e != nil {
		b.Log(e)
	}
	for {
		select {
		case ruleI := <-b.inrule:
			if pinStr != "" {
				err = hwio.ClosePin(pin)
				b.Error(err)
			}
			pinStr, err = util.ParseString(ruleI, "Pin")
			if err != nil {
				b.Error(err)
				continue
			}
			pin, err = hwio.GetPin(pinStr)
			if err != nil {
				b.Error(err)
				continue
			}
			err = hwio.PinMode(pin, hwio.INPUT)
			if err != nil {
				b.Error(err)
				continue
			}
		case <-b.quit:
			// quit the block
			err = hwio.ClosePin(pin)
			b.Error(err)
			return
		case c := <-b.queryrule:
			// deal with a query request
			c <- map[string]interface{}{
				"Pin": pinStr,
			}

		case <-b.inpoll:
			if pin == 0 {
				continue
			}
			v, err := hwio.AnalogRead(pin)
			if err != nil {
				b.Error(err)
				continue
			}
			out := map[string]interface{}{
				"value": float64(v),
				"pin":   pinStr,
			}
			b.out <- out
		}
	}
}
