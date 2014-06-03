package library

import (
	"errors"

	"github.com/nytlabs/gojee"
	"github.com/nytlabs/streamtools/st/blocks" // blocks
)

// specify those channels we're going to use to communicate with streamtools
type Map struct {
	blocks.Block
	queryrule chan blocks.MsgChan
	inrule    blocks.MsgChan
	in        blocks.MsgChan
	out       blocks.MsgChan
	quit      blocks.MsgChan
}

func setVal(m interface{}, key string, val interface{}) error {
	min, ok := m.(map[string]interface{})
	if !ok {
		return errors.New("type assertion failed")
	}
	min[key] = val
	return nil
}

// parse and lex each key
func parseKeys(mapRule map[string]interface{}) (interface{}, error) {
	t := make(map[string]interface{})

	for k, e := range mapRule {
		switch r := e.(type) {
		case map[string]interface{}:
			j, err := parseKeys(r)
			if err != nil {
				return nil, err
			}
			setVal(t, k, j)
		case string:
			lexed, err := jee.Lexer(r)
			if err != nil {
				return nil, err
			}
			tree, err := jee.Parser(lexed)
			if err != nil {
				return nil, err
			}
			setVal(t, k, tree)
		}
	}

	return t, nil
}

// run jee.eval for each key
func evalMap(mapRule map[string]interface{}, msg map[string]interface{}) (map[string]interface{}, error) {
	nt := make(map[string]interface{})
	for k, _ := range mapRule {
		switch c := mapRule[k].(type) {
		case *jee.TokenTree:
			e, err := jee.Eval(c, msg)
			if err != nil {
				return nil, err
			}
			setVal(nt, k, e)
		case map[string]interface{}:
			e, err := evalMap(c, msg)
			if err != nil {
				return nil, err
			}
			msg[k] = setVal(nt, k, e)
		}
	}
	return nt, nil
}

// recursively copy map
func recCopy(msg map[string]interface{}) map[string]interface{} {
	n := make(map[string]interface{})

	for k, _ := range msg {
		switch m := msg[k].(type) {
		case map[string]interface{}:
			setVal(n, k, recCopy(m))
		default:
			setVal(n, k, m)
		}
	}
	return n
}

// we need to build a simple factory so that streamtools can make new blocks of this kind
func NewMap() blocks.BlockInterface {
	return &Map{}
}

// Setup is called once before running the block. We build up the channels and specify what kind of block this is.
func (b *Map) Setup() {
	b.Kind = "Core"
	b.Desc = "maps inbound data onto outbound data, providing a way to restructure or rename elements"
	b.in = b.InRoute("in")
	b.inrule = b.InRoute("rule")
	b.queryrule = b.QueryRoute("rule")
	b.quit = b.Quit()
	b.out = b.Broadcast()
}

// Run is the block's main loop. Here we listen on the different channels we set up.
func (b *Map) Run() {
	additive := true
	mapRule := map[string]interface{}{}
	parsed, _ := parseKeys(mapRule)

	for {
		select {
		case ruleI := <-b.inrule:
			// set a parameter of the block
			rule, ok := ruleI.(map[string]interface{})
			if !ok {
				b.Error(errors.New("could not assert rule to map[string]interface{}"))
			}
			additiveStr, ok := rule["Additive"]
			if !ok {
				// 'Additive' wasn't set in the rule, but it defaults to true
				additiveStr = true
			}
			additive, ok = additiveStr.(bool)
			mapRuleI, ok := rule["Map"]
			if !ok {
				b.Error(errors.New("could not find Map in rule"))
				break
			}
			mapRule = mapRuleI.(map[string]interface{})
			p, err := parseKeys(mapRule)
			if err == nil {
				parsed = p
			} else {
				b.Error(err)
			}
		case <-b.quit:
			// quit the block
			return
		case msg := <-b.in:
			// deal with inbound data
			if parsed == nil {
				continue
			}
			result := make(map[string]interface{})
			if additive == true {
				result = recCopy(msg.(map[string]interface{}))
			}

			in := msg.(map[string]interface{})
			evaled, err := evalMap(parsed.(map[string]interface{}), in)
			if err != nil {
				b.Error(err)
			}

			for k, _ := range evaled {
				result[k] = evaled[k]
			}
			b.out <- result

		case MsgChan := <-b.queryrule:
			// deal with a query request
			rule := map[string]interface{}{
				"Map":      mapRule,
				"Additive": additive,
			}
			MsgChan <- rule
		}
	}
}
