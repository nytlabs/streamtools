package library

import (
	"github.com/nytlabs/gojee"
	"github.com/nytlabs/streamtools/st/blocks" // blocks
	"log"
)

type Filter struct {
	blocks.Block
	queryrule chan chan interface{}
	inrule    chan interface{}
	in        chan interface{}
	out       chan interface{}
	quit      chan interface{}
	filter    string
}

// a bit of boilerplate for streamtools
func NewFilter() blocks.BlockInterface {
	return &Filter{}
}

func (b *Filter) Setup() {
	b.Kind = "Filter"
	b.in = b.InRoute("in")
	b.inrule = b.InRoute("rule")
	b.queryrule = b.QueryRoute("rule")
	b.quit = b.Quit()
	b.out = b.Broadcast()
}

func (b *Filter) Run() {

	var parsed *jee.TokenTree

	for {
		select {
		case msg := <-b.in:
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
				b.out <- map[string]interface{}{
					"Msg": msg,
				}
			}

		case ruleI := <-b.inrule:
			rule := ruleI.(map[string]string)
			lexed, err := jee.Lexer(rule["Filter"])
			if err != nil {
				log.Println(err)
				break
			}

			tree, err := jee.Parser(lexed)
			if err != nil {
				log.Println(err)
				break
			}

			parsed = tree
			b.filter = rule["Filter"]

		case c := <-b.queryrule:
			// deal with a query request
			c <- map[string]string{
				"Filter": b.filter,
			}
		case <-b.quit:
			// quit the block
			return
		}
	}
}
