package library

import (
	"github.com/nytlabs/gojee"
	"github.com/nytlabs/streamtools/st/blocks" // blocks
	"github.com/nytlabs/streamtools/st/util"
)

type Filter struct {
	blocks.Block
	queryrule chan blocks.MsgChan
	inrule    blocks.MsgChan
	in        blocks.MsgChan
	out       blocks.MsgChan
	quit      blocks.MsgChan
}

// a bit of boilerplate for streamtools
func NewFilter() blocks.BlockInterface {
	return &Filter{}
}

func (b *Filter) Setup() {
	b.Kind = "Core"
	b.Desc = "selectively emits messages based on criteria defined in this block's rule"
	b.in = b.InRoute("in")
	b.inrule = b.InRoute("rule")
	b.queryrule = b.QueryRoute("rule")
	b.quit = b.Quit()
	b.out = b.Broadcast()
}

func (b *Filter) Run() {
	filter := ". != null"
	lexed, _ := jee.Lexer(filter)
	parsed, _ := jee.Parser(lexed)

	for {
		select {
		case msg := <-b.in:
			if parsed == nil {
				b.Error("no filter set")
				break
			}

			e, err := jee.Eval(parsed, msg)
			if err != nil {
				b.Error(err)
				break
			}

			eval, ok := e.(bool)
			if !ok {
				break
			}

			if eval == true {
				b.out <- msg
			}

		case ruleI := <-b.inrule:
			filterS, err := util.ParseString(ruleI, "Filter")
			if err != nil {
				b.Error("bad filter")
				break
			}

			lexed, err := jee.Lexer(filterS)
			if err != nil {
				b.Error(err)
				break
			}

			tree, err := jee.Parser(lexed)
			if err != nil {
				b.Error(err)
				break
			}

			parsed = tree
			filter = filterS

		case c := <-b.queryrule:
			// deal with a query request
			c <- map[string]interface{}{
				"Filter": filter,
			}
		case <-b.quit:
			// quit the block
			return
		}
	}
}
