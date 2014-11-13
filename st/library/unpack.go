package library

import (
	"errors"

	"github.com/nytlabs/gojee"                 // jee
	"github.com/nytlabs/streamtools/st/blocks" // blocks
	"github.com/nytlabs/streamtools/st/util"
)

type Unpack struct {
	blocks.Block
	queryrule chan blocks.MsgChan
	inrule    blocks.MsgChan
	in        blocks.MsgChan
	out       blocks.MsgChan
	quit      blocks.MsgChan
}

func NewUnpack() blocks.BlockInterface {
	return &Unpack{}
}

func (b *Unpack) Setup() {
	b.Kind = "Core"
	b.Desc = "splits an array of objects from incoming data, emitting each element as a separate message"
	b.in = b.InRoute("in")
	b.inrule = b.InRoute("rule")
	b.queryrule = b.QueryRoute("rule")
	b.quit = b.Quit()
	b.out = b.Broadcast()
}

func (b *Unpack) Run() {
	var arrayPath, labelPath string
	var err error
	var arrayTree, labelTree *jee.TokenTree
	var label interface{}

	for {
		select {
		case ruleI := <-b.inrule:
			rule, ok := ruleI.(map[string]interface{})
			if !ok {
				b.Error(errors.New("cannot assert rule to map"))
			}

			arrayPath, err = util.ParseString(rule, "ArrayPath")
			if err != nil {
				b.Error(err)
				continue
			}

			labelPath, err = util.ParseString(rule, "LabelPath")
			if err != nil {
				b.Error(err)
				continue
			}

			arrayToken, err := jee.Lexer(arrayPath)
			if err != nil {
				b.Error(err)
				continue
			}

			arrayTree, err = jee.Parser(arrayToken)
			if err != nil {
				b.Error(err)
				continue
			}

			if labelPath == arrayPath {
				b.Error(errors.New("cannot label unpacked objects with the original array"))
				continue
			}

			labelToken, err := jee.Lexer(labelPath)
			if err != nil {
				b.Error(err)
				continue
			}

			labelTree, err = jee.Parser(labelToken)
			if err != nil {
				b.Error(err)
				continue
			}

		case <-b.quit:
			// quit the block
			return
		case msg := <-b.in:
			if arrayTree == nil {
				continue
			}

			arrInterface, err := jee.Eval(arrayTree, msg)
			if err != nil {
				b.Error(err)
				continue
			}

			arr, ok := arrInterface.([]interface{})
			if !ok {
				b.Error(errors.New("cannot assert " + arrayPath + " to array"))
				continue
			}

			if labelTree != nil {
				label, err = jee.Eval(labelTree, msg)
				if err != nil {
					b.Error(err)
					continue
				}
			}

			for _, out := range arr {
				if labelPath == "" {
					b.out <- out
					continue
				}

				outMap := make(map[string]interface{})
				outMap["Value"] = out
				outMap["Label"] = label
				b.out <- outMap
			}
		case c := <-b.queryrule:
			// deal with a query request
			out := map[string]interface{}{
				"ArrayPath": arrayPath,
				"LabelPath": labelPath,
			}
			c <- out
		}
	}
}
