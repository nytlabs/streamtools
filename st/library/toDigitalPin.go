// +build arm

package library

import (
	"errors"
	"log"

	"github.com/mrmorphic/hwio"                // hwio
	"github.com/nytlabs/gojee"                 // jee
	"github.com/nytlabs/streamtools/st/blocks" // blocks
	"github.com/nytlabs/streamtools/st/util"
)

type ToDigitalPin struct {
	blocks.Block
	queryrule chan blocks.MsgChan
	inrule    blocks.MsgChan
	in        blocks.MsgChan
	quit      blocks.MsgChan
}

func NewToDigitalPin() blocks.BlockInterface {
	return &ToDigitalPin{}
}

// Setup is called once before running the block. We build up the channels and specify what kind of block this is.
func (b *ToDigitalPin) Setup() {
	b.Kind = "Hardware I/O"
	b.Desc = "(embedded applications) sets the state of a digital pin"
	b.inrule = b.InRoute("rule")
	b.in = b.InRoute("in")
	b.queryrule = b.QueryRoute("rule")
	b.quit = b.Quit()
}

// Run is the block's main loop. Here we listen on the different channels we set up.
func (b *ToDigitalPin) Run() {
	var pin hwio.Pin
	var pinStr string
	var tree *jee.TokenTree
	var path string
	var err error
	for {
		select {
		case ruleI := <-b.inrule:
			path, err = util.ParseString(ruleI, "Path")
			if err != nil {
				b.Error(err)
				continue
			}
			token, err := jee.Lexer(path)
			if err != nil {
				b.Error(err)
				continue
			}
			tree, err = jee.Parser(token)
			if err != nil {
				b.Error(err)
				continue
			}
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
			err = hwio.PinMode(pin, hwio.OUTPUT)
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
				"Pin":  pinStr,
				"Path": path,
			}
		case msg := <-b.in:
			if tree == nil {
				continue
			}
			valI, err := jee.Eval(tree, msg)
			if err != nil {
				b.Error(err)
				continue
			}
			val, ok := valI.(float64)
			if !ok {
				log.Println(msg)
				b.Error(errors.New("couldn't assert value to a float"))
				continue
			}
			if int(val) == 0 {
				hwio.DigitalWrite(pin, hwio.LOW)
			} else if int(val) == 1 {
				hwio.DigitalWrite(pin, hwio.HIGH)
			} else {
				b.Error(errors.New("value must be 0 for LOW and 1 for HIGH"))
				continue
			}

		}
	}
}
