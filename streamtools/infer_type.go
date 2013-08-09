package streamtools

import (
	"github.com/bitly/go-simplejson"
	"log"
	"reflect"
)

// flattenType returns a map of flatten keys of the incoming dictionary, and
// values as the corresponding JSON types.
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

// convertStringMapToJson simply takes a map of strings to strings,
// and converts it to a simplejson.Json object.
func convertStringMapToJson(m map[string]string) *simplejson.Json {
	msg, _ := simplejson.NewJson([]byte("{}"))
	for k, v := range m {
		msg.Set(k, v)
	}
	return msg
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
			msg := convertStringMapToJson(flat)
			outChan <- *msg
		}
	}
}
