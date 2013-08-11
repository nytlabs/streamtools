package streamtools

import (
	"github.com/bitly/go-simplejson"
	"log"
)

func calcDiff(p map[string]interface{}, n map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{})
	for k, v := range p {
		switch pv := v.(type) {
		case map[string]interface{}:
			// got an object
			nv, _ := n[k].(map[string]interface{})
			out[k] = calcDiff(pv, nv)
		case []interface{}:
			// got an array
			// TODO: there's no such thing as an arithmetic difference
			// between two arrays; there is, however, difference as
			// defined in set theory, which is what I've used here.
			// Does it make sense to mix that with arithmetic?
			// What about set.symmetric_difference?
			nv, _ := n[k].([]interface{})
			nn := NewSetFromSlice(nv)
			d := nn.Difference(NewSetFromSlice(pv))
			out[k] = d.ToSlice()
		case int, float32, float64:
			// got a number
			nv, _ := n[k].(float64)
			switch pv := pv.(type) {
			case int:
				out[k] = nv - float64(pv)
			case float32:
				out[k] = nv - float64(pv)
			case float64:
				out[k] = nv - pv
			}
		default:
			// nil, string, bool; do nothing.
		}
	}
	return out
}

var Diff TransferFunction = func(inChan chan simplejson.Json, outChan chan simplejson.Json) {
	var prev map[string]interface{} = nil
	for {
		select {
		case m := <-inChan:
			blob, err := m.Map()
			if err != nil {
				log.Fatalln(err)
			}
			if prev != nil {
				diff := calcDiff(prev, blob)
				log.Println(diff)
				msg := convertInterfaceMapToSimplejson(diff)
				outChan <- *msg
			}
			prev = blob
		}
	}
}
