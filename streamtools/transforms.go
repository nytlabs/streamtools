package streamtools

import (
	"github.com/bitly/go-simplejson"
	"log"
	"reflect"
)

// calcDiff returns a map of the difference between the two given maps, where 
// numerical values are return arithmetic diff values, slices return the result 
// of a set difference operation, and all other values (string, bool, null) 
// don't return a diff.
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

// Diff reads from an incoming JSON channel, and calculates the difference
// between consecutive messages, which is then put on an outgoing channel;
// this transfer block outputs n-1 messages for every n input messages.
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

// flattenType returns a map of flatten keys of the incoming dictionary, and
// values as the corresponding JSON types.
// TODO: de-clunkify, cf. calcDiff.
func flattenType(d map[string]interface{}, p string) map[string]string {
	out := make(map[string]string)
	for key, value := range d {
		new_p := ""
		if len(p) > 0 {
			new_p = p + "." + key
		} else {
			new_p = key
		}
		if value == nil {
			// got JSON type null
			out[key] = "null"
		} else if reflect.TypeOf(value).Kind() == reflect.Map {
			// got an object
			s, ok := value.(map[string]interface{})
			if ok {
				for k, v := range flattenType(s, new_p) {
					out[k] = v
				}
			} else {
				log.Fatalf("expected type map, got something else instead. key=%s, s=%s", key, s)
			}
		} else if reflect.TypeOf(value).Kind() == reflect.Slice {
			// got an array
			new_p += ".[]"
			s, ok := value.([]interface{})
			if ok {
				for _, d2 := range s {
					if reflect.TypeOf(d2).Kind() == reflect.Map {
						s2, ok := d2.(map[string]interface{})
						if ok {
							for k, v := range flattenType(s2, new_p) {
								out[k] = v
							}
						} else {
							log.Fatalf("expected type map, got something else instead. key=%s, s2=%s", key, s2)
						}
					} else {
						// array here contains non-objects, so just save element type and break
						// note JSON doesn't require arrays have uniform type, but we'll assume it does
						out[key] = "Array[ " + prettyPrintJsonType(d2) + " ]"
						break
					}
				}
			} else {
				log.Fatalf("expected type []interface{}, got something else instead. key=%s, s=%s", key, s)
			}
		} else {
			// got a basic type: Number, Boolean, or String
			out[new_p] = prettyPrintJsonType(value)
		}
	}
	return out
}

// prettyPrintJsonType accepts a variable (of type interface{}) and
// returns a human-readable string of "Number", "Boolean", "String", or "UNKNOWN".
func prettyPrintJsonType(value interface{}) string {
	switch t := value.(type) {
	case float64:
		return "Number"
	case bool:
		return "Boolean"
	case string:
		return "String"
	default:
		log.Fatalf("unexpected type %T", t)
	}
	return "UNKNOWN"
}

// InferType reads from an incoming channel msgChan, flattens and
// types the event, and puts it on another channel outChan.
var InferType TransferFunction = func(inChan chan simplejson.Json, outChan chan simplejson.Json) {
	for {
		select {
		case m := <-inChan:
			blob, err := m.Map()
			if err != nil {
				log.Fatalln(err)
			}
			flat := flattenType(blob, "")
			msg := convertStringMapToSimplejson(flat)
			outChan <- *msg
		}
	}
}
