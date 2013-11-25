package blocks

import (
	"encoding/json"
	"strconv"
	"strings"
)

// getKeyValues returns values for all paths, including arrays
// {"foo":"bar"} returns [bar] for string "foo"
// {"foo":["bar","bar","bar"]} returns [bar, bar, bar] for string "foo[]"
// {"foo":[{"type":"bar"},{"type":"baz"}]} returns [bar, baz] for string "foo[].type"
// {"foo":["bar","baz"]} returns [bar] for string "foo[0]"

// this function is obscene :(
func getKeyValues(d interface{}, p string) []interface{} {
	var values []interface{}
	var key string
	var rest string

	keyIdx := strings.Index(p, ".")
	brkIdx := strings.Index(p, "[")

	if keyIdx != -1 {
		key = p[:keyIdx]
		rest = p[keyIdx+1:]
	} else {
		key = p
	}

	if brkIdx != -1 && brkIdx != 0 {
		key = p[:brkIdx]
		rest = p[brkIdx:]
	}

	bStart := strings.Index(key, "[")
	bEnd := strings.Index(key, "]")
	var id int64
	id = -1
	if bStart == 0 && bEnd != 1 {
		id, _ = strconv.ParseInt(key[bStart+1:bEnd], 10, 64)
	}

	switch d := d.(type) {
	case map[string]interface{}:
		if len(rest) > 0 {
			x := getKeyValues(d[key], rest)
			for _, z := range x {
				values = append(values, z)
			}
		} else {
			_, ok := (d[p]).([]interface{})
			if ok == false {
				values = append(values, d[p])
			}
		}
	case []int:
		var ids []int64
		if id != -1 {
			if len(d) == 0 || int(id) >= len(d) || id < 0 {
				break
			}
			ids = append(ids, int64(id))
		} else {
			for i := range d {
				ids = append(ids, int64(i))
			}
		}

		for _, id := range ids {
			values = append(values, d[id])
		}
	case []string:
		var ids []int64
		if id != -1 {
			if len(d) == 0 || int(id) >= len(d) || id < 0 {
				break
			}

			ids = append(ids, int64(id))
		} else {
			for i := range d {
				ids = append(ids, int64(i))
			}
		}

		for _, id := range ids {
			values = append(values, d[id])
		}
	case []bool:
		var ids []int64
		if id != -1 {
			if len(d) == 0 || int(id) >= len(d) || id < 0 {
				break
			}
			ids = append(ids, int64(id))
		} else {
			for i := range d {
				ids = append(ids, int64(i))
			}
		}

		for _, id := range ids {
			values = append(values, d[id])
		}
	case []float64:
		var ids []int64
		if id != -1 {
			if len(d) == 0 || int(id) >= len(d) || id < 0 {
				break
			}
			ids = append(ids, int64(id))
		} else {
			for i := range d {
				ids = append(ids, int64(i))
			}
		}

		for _, id := range ids {
			values = append(values, d[id])
		}
	case []interface{}:
		var ids []int64
		if id != -1 {
			if len(d) == 0 || int(id) >= len(d) || id < 0 {
				break
			}
			ids = append(ids, int64(id))
		} else {
			for i := range d {
				ids = append(ids, int64(i))
			}
		}

		for _, id := range ids {
			if len(rest) > 0 {
				x := getKeyValues(d[id], rest)
				for _, z := range x {
					values = append(values, z)
				}
			} else {
				values = append(values, d[id])
			}
		}

	default:
	}

	return values
}

func equals(value interface{}, comparator interface{}) bool {
	switch value := value.(type) {
	case json.Number:
		// not sure about comparing floats
		v, err := value.Float64()
		if err != nil {
			return false
		}
		c, ok := comparator.(float64)
		if ok == false {
			return false
		}
		return v == c
	case string:
		return value == comparator
	case bool:
		return value == comparator
	default:
		if value == nil && comparator == nil {
			return true
		}
		return false
	}
}

func greaterthan(value interface{}, comparator interface{}) bool {
	switch value := value.(type) {
	case json.Number:
		// not sure about comparing floats
		v, err := value.Float64()
		if err != nil {
			return false
		}
		c, ok := comparator.(float64)
		if ok == false {
			return false
		}
		return v > c
	default:
		return false
	}
}

func lessthan(value interface{}, comparator interface{}) bool {
	switch value := value.(type) {
	case json.Number:
		// not sure about comparing floats
		v, err := value.Float64()
		if err != nil {
			return false
		}
		c, ok := comparator.(float64)
		if ok == false {
			return false
		}
		return v < c
	default:
		return false
	}
}

func subsetof(value interface{}, comparator interface{}) bool {
	switch value := value.(type) {
	case string:
		return strings.Contains(value, comparator.(string))
	}
	return false
}
