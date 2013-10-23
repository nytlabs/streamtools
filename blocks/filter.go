package blocks

import (
	"github.com/bitly/go-simplejson"
	"log"
	"strings"
)

type filterRule struct {
	op     string
	params interface{}
}

type filterFunc func(*filterRule, *simplejson.Json) *simplejson.Json

// identity filter
type identityRule struct{}

func identity(r *filterRule, msg *simplejson.Json) *simplejson.Json {
	return msg
}

// key in object
type keyInObjectRule struct {
	key string
}

func keyInObject(r *filterRule, msg *simplejson.Json) *simplejson.Json {
	thisRule := r.params.(keyInObjectRule)
	if _, ok := msg.CheckGet(thisRule.key); ok {
		return msg
	} else {
		return nil
	}
}

// value greater than
type valueCompareRule struct {
	key       string
	condition string
	limit     interface{}
}

func valueCompare(r *filterRule, msg *simplejson.Json) *simplejson.Json {
	thisRule := r.params.(valueCompareRule)
	value := msg.Get(thisRule.key).Interface()
	switch value := value.(type) {
	default:
		log.Fatal("unsupported type for greater than comparison")
	case float64:
		limit, ok := thisRule.limit.(float64)
		if !ok {
			log.Fatal("comparisons must be of the same type")
		}
		switch thisRule.condition {
		case ">":
			if value > limit {
				return msg
			}
		case "<":
			if value < limit {
				return msg
			}
		case "=":
			if value == limit {
				return msg
			}
		}
	}
	return nil
}

// array length comparisons
type arrayLenRule struct {
	condition string
	value     int
	path      string
}

func arrayLen(r *filterRule, msg *simplejson.Json) *simplejson.Json {
	thisRule := r.params.(arrayLenRule)
	path := strings.Split(thisRule.path, ".")

	val, err := msg.GetPath(path...).Array()
	if err != nil {
		log.Println(err.Error())
	}

	switch thisRule.condition {
	case ">":
		if len(val) > thisRule.value {
			return msg
		}
	case "<":
		if len(val) < thisRule.value {
			return msg
		}
	case "=":
		if len(val) == thisRule.value {
			return msg
		}
	}

	return nil
}

func Filter(b *Block) {

	var filter filterFunc

	rule := &filterRule{}
	unmarshal(<-b.Routes["set_rule"], &rule)

	switch rule.op {
	case "keyInObject":
		filter = keyInObject
	case "valueCompare":
		filter = valueCompare
	case "arrayLen":
		filter = arrayLen
	case "identity":
		filter = identity
	default:
		log.Fatal("unknown filter", rule.op)
	}

	for {
		select {
		case msg := <-b.InChan:
			out := filter(rule, msg)
			if out != nil {
				broadcast(b.OutChans, out)
			}
		case msg := <-b.AddChan:
			updateOutChans(msg, b)
		}
	}
}
