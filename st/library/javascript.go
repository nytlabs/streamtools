package library

import (
	"github.com/nytlabs/streamtools/st/blocks" // blocks
	"github.com/nytlabs/streamtools/st/util"
	"github.com/robertkrimen/otto"
	_ "github.com/robertkrimen/otto/underscore"
)

// specify those channels we're going to use to communicate with streamtools
type Javascript struct {
	blocks.Block
	queryrule chan blocks.MsgChan
	inrule    blocks.MsgChan
	inpoll    blocks.MsgChan
	in        blocks.MsgChan
	out       blocks.MsgChan
	quit      blocks.MsgChan
}

// we need to build a simple factory so that streamtools can make new blocks of this kind
func NewJavascript() blocks.BlockInterface {
	return &Javascript{}
}

// Setup is called once before running the block. We build up the channels and specify what kind of block this is.
func (b *Javascript) Setup() {
	b.Kind = "Core"
	b.Desc = "transform messages with javascript (includes underscore.js)"
	b.in = b.InRoute("in")
	b.inrule = b.InRoute("rule")
	b.queryrule = b.QueryRoute("rule")
	b.quit = b.Quit()
	b.out = b.Broadcast()
}

// Run is the block's main loop. Here we listen on the different channels we set up.
func (b *Javascript) Run() {
	messageIn := "input"
	messageOut := "output"
	script := `output = input`

	vm := otto.New()
	program, _ := vm.Compile("javascript", script)

	for {
		select {
		case ruleI := <-b.inrule:
			tmpMin, err := util.ParseString(ruleI, "MessageIn")
			if err != nil {
				b.Error(err)
				break
			}

			tmpMout, err := util.ParseString(ruleI, "MessageOut")
			if err != nil {
				b.Error(err)
				break
			}

			tmpScript, err := util.ParseString(ruleI, "Script")
			if err != nil {
				b.Error(err)
				break
			}

			tmpProgram, err := vm.Compile("javascript", tmpScript)
			if err != nil {
				b.Error(err)
				break
			}

			messageIn = tmpMin
			messageOut = tmpMout
			script = tmpScript
			program = tmpProgram

		case <-b.quit:
			// quit the block
			return
		case m := <-b.in:
			if program == nil {
				break
			}

			err := vm.Set(messageOut, map[string]interface{}{})
			if err != nil {
				b.Error(err)
				break
			}

			err = vm.Set(messageIn, m)
			if err != nil {
				b.Error(err)
				break
			}

			_, err = vm.Run(program)
			if err != nil {
				b.Error(err)
				break
			}

			g, err := vm.Get(messageOut)
			if err != nil {
				b.Error(err)
				break
			}
			o, err := g.Export()
			if err != nil {
				b.Error(err)
				break
			}

			b.out <- o
		case c := <-b.queryrule:
			c <- map[string]interface{}{
				"MessageIn":  messageIn,
				"MessageOut": messageOut,
				"Script":     script,
			}
		}
	}
}
