package blocks

import (
	"log"
	"regexp"
	"strconv"
	"strings"
)

// getKeyValues returns values for all paths, including arrays
// {"foo":"bar"} returns [bar] for string "foo"
// {"foo":["bar","bar","bar"]} returns [bar, bar, bar] for string "foo[]"
// {"foo":[{"type":"bar"},{"type":"baz"}]} returns [bar, baz] for string "foo[].type"
// {"foo":["bar","baz"]} returns [bar] for string "foo[0]"
//
// getKeyValues also supports bracket access in case your keys include periods.
// {"key.includes.periods":1} returns [1] for string "['key.includes.periods']"
// {"foo":{"bar.bar":{"baz":1}} returns [1] for string 'foo["bar.bar"].baz'
// {"foo.bar"{"baz":{"bing.bong":{"boo":1}}}} returns [1] for string '["foo.bar"].baz["bing.bong"]["boo"]'

// this function is obscene :(
func getKeyValues(d interface{}, p string) []interface{} {
	var values []interface{}
	var key string
	var rest string
	var id int64
	id = -1

	keyIdx := strings.Index(p, ".")
	brkIdx := strings.Index(p, "[")
	escIdx := strings.Index(p, "[\"")

	if escIdx == -1 {
		escIdx = strings.Index(p, "['")
	}

	if escIdx == 0 {
		endescIdx := strings.Index(p, "\"]")
		if endescIdx == -1 {
			endescIdx = strings.Index(p, "']")
		}

		key = p[escIdx+2 : endescIdx]
		rest = p[endescIdx+2:]

		if len(rest) > 0 && rest[0] == '.' {
			rest = rest[1:]
		}
	} else {

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

		if bStart == 0 && bEnd != 1 {
			id, _ = strconv.ParseInt(key[bStart+1:bEnd], 10, 64)
		}

	}

	switch d := d.(type) {
	case map[string]interface{}:
		if len(rest) > 0 {
			x := getKeyValues(d[key], rest)
			for _, z := range x {
				values = append(values, z)
			}
		} else {
			_, ok := d[key]
			if ok {
				_, ok := (d[key]).([]interface{})
				if ok == false {
					values = append(values, d[key])
				}
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
	case float64:
		// comparing floats.....?
		c, ok := comparator.(float64)
		if ok == false {
			return false
		}
		return value == c
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
	case float64:
		c, ok := comparator.(float64)
		if ok == false {
			return false
		}
		return value > c
	default:
		return false
	}
}

func lessthan(value interface{}, comparator interface{}) bool {
	switch value := value.(type) {
	case float64:
		c, ok := comparator.(float64)
		if ok == false {
			return false
		}
		return value < c
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

func regexmatch(value interface{}, comparator interface{}) bool {
	r, ok := comparator.(*regexp.Regexp)
	if !ok {
		log.Println("non regex passed into regexmatch")
		return false
	}
	switch value := value.(type) {
	case string:
		return r.Match([]byte(value))
	}
	return false
}

func keyin(value interface{}, comparator interface{}) bool {
	return true
}
