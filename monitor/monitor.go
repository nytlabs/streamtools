package main

import (
	"flag"
	"github.com/bitly/go-simplejson"
	"log"
	"reflect"
)

var (
	json = flag.String("json", "{\"a\":6.2, \"b\":{ \"c\": 3, \"g\": { \"h\": \"POO\"} }, \"d\": \"catz\", \"e\":[14,15], \"f\": 8.0, \"i\": [{ \"j\": \"poop\"},{ \"j\": \"pop\",\"k\": true }], \"m\": null, \"n\": [\"egg\",\"chicken\",\"gravy\" ] }", "a blob")
)

func Flatten(d map[string]interface{}, p string) map[string]interface{} {
	out := make(map[string]interface{})
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
				for k, v := range Flatten(s, new_p) {
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
							for k, v := range Flatten(s2, new_p) {
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

func main() {

	flag.Parse()
	log.Printf("")
	log.Printf("json=%s", *json)

	blob, err := simplejson.NewJson([]byte(*json))
	if err != nil {
		log.Fatalf("error converting string to Json: %s", err)
	}
	log.Println("")

	mblob, err := blob.Map()
	if err != nil {
		log.Fatalln(err)
	}

	poo := Flatten(mblob, "")
	for k, v := range poo {
		log.Printf("k= %v\tt= %v", k, v)
	}
}
