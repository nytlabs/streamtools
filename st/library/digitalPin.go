// +build arm

package library

import (
	"github.com/mrmorphic/hwio"                // hwio
	"github.com/nytlabs/streamtools/st/blocks" // blocks
	"github.com/nytlabs/streamtools/st/util"
)

type DigitalPin struct {
	blocks.Block
	queryrule chan blocks.MsgChan
	inrule    blocks.MsgChan
	inpoll    blocks.MsgChan
	out       blocks.MsgChan
	quit      blocks.MsgChan
}

func NewDigitalPin() blocks.BlockInterface {
	return &DigitalPin{}
}

// Setup is called once before running the block. We build up the channels and specify what kind of block this is.
func (b *DigitalPin) Setup() {
	b.Kind = "Hardware I/O"
	b.Desc = "(embedded applications) returns current state of the digital pin"
	b.inrule = b.InRoute("rule")
	b.inpoll = b.InRoute("poll")
	b.queryrule = b.QueryRoute("rule")
	b.quit = b.Quit()
	b.out = b.Broadcast()
}

// Run is the block's main loop. Here we listen on the different channels we set up.
func (b *DigitalPin) Run() {
	var pin hwio.Pin
	var pinStr string
	var err error
	for {
		select {
		case ruleI := <-b.inrule:
			if pinStr != "" {
				b.Log("closing pin " + pinStr)
				err = hwio.ClosePin(pin)
				if err != nil {
					b.Error(err)
				}
			}
			pinStr, err = util.ParseString(ruleI, "Pin")
			if err != nil {
				b.Error(err)
				continue
			}
			pin, err = hwio.GetPin(pinStr)
			if err != nil {
				pinStr = ""
				pin = 0
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
			v, err := hwio.DigitalRead(pin)
			if err != nil {
				b.Log(v)
				b.Error(err)
				continue
			}
			outValue := float64(v)
			out := map[string]interface{}{
				"value": outValue,
				"pin":   pinStr,
			}
			b.out <- out
		}
	}
}
