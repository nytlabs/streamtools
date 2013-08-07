package main

import (
    "flag"
)

var (
    mask = flag.String("mask", "", "JSON describing keys to pass")
)

// recursively removes keys for JSON object
func filter(mask map[string]interface{}, input map[string]interface{}) map[string]interface{} {
	t := make(map[string]interface{})

	// iterate only over mask keys
	for k, v := range mask {
		switch vv := input[k].(type) {
		case []interface{}: // arrays
            a := make([]interface{}, len(vv))
			for i, u := range vv {
				a[i] = filter(mask[k].(map[string]interface{}), u.(map[string]interface{}))
			}
            t[k] = a
		case map[string]interface{}: // object
			t[k] = filter(mask[k].(map[string]interface{}), input[k].(map[string]interface{}))
		default: // string, int, float
			_, ok := input[k]
			if ok && v == nil {
				t[k] = input[k]
			}
		}
	}
	return t
}
