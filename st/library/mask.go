package library

import (
	"github.com/nytlabs/streamtools/st/blocks" // blocks
)

// specify those channels we're going to use to communicate with streamtools
type Mask struct {
	blocks.Block
	queryrule chan blocks.MsgChan
	inrule    blocks.MsgChan
	in        blocks.MsgChan
	out       blocks.MsgChan
	quit      blocks.MsgChan
}

// a bit of boilerplate for streamtools
func NewMask() blocks.BlockInterface {
	return &Mask{}
}

func (b *Mask) Setup() {
	b.Kind = "Core"
	b.Desc = "emits a subset of the inbound message by specifying the desired JSON output structure in this block's rule"
	b.in = b.InRoute("in")
	b.inrule = b.InRoute("rule")
	b.queryrule = b.QueryRoute("rule")
	b.quit = b.Quit()
	b.out = b.Broadcast()
}

func maskJSON(maskMap map[string]interface{}, input map[string]interface{}) map[string]interface{} {
	t := make(map[string]interface{})

	if len(maskMap) == 0 {
		return input
	}

	for k, _ := range maskMap {
		val, ok := input[k]
		if ok {
			switch v := val.(type) {
			case map[string]interface{}:
				maskNext, ok := maskMap[k].(map[string]interface{})
				if ok {
					t[k] = maskJSON(maskNext, v)
				} else {
					t[k] = v
				}
			default:
				t[k] = val
			}
		}
	}

	return t
}

// Mask modifies a JSON stream with an additive key filter. Mask uses the JSON
// object recieved through the rule channel to determine which keys should be
// included in the resulting object. An empty JSON object ({}) is used as the
// notation to include all values for a key.
//
// For instance, if the JSON rule is:
//        {"a":{}, "b":{"d":{}},"x":{}}
// And an incoming message looks like:
//        {"a":24, "b":{"c":"test", "d":[1,3,4]}, "f":5, "x":{"y":5, "z":10}}
// The resulting object after the application of Mask would be:
//        {"a":24, "b":{"d":[1,3,4]}, "x":{"y":5, "z":10}}
func (b *Mask) Run() {
	mask := make(map[string]interface{})

	for {
		select {
		case ruleI := <-b.inrule:
			rule := ruleI.(map[string]interface{})
			if tmp, ok := rule["Mask"].(map[string]interface{}); ok {
				mask = tmp
			}
		case c := <-b.queryrule:
			c <- map[string]interface{}{
				"Mask": mask,
			}
		case msg := <-b.in:
			msgMap, msgOk := msg.(map[string]interface{})
			if msgOk {
				b.out <- maskJSON(mask, msgMap)
			}
		case <-b.quit:
			// quit the block
			return
		}
	}
}
