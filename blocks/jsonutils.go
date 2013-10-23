package blocks

import (
	"github.com/bitly/go-simplejson"
	"log"
)

func getPath(msg *simplejson.Json, path string) interface{} {
	return msg.Get(path).Interface()
}

func equals(value interface{}, comparator interface{}) bool {
	switch value := value.(type) {
	case int:
		c := comparator.(float64)
		return value == int(c)
	case string:
		return value == comparator
	case bool:
		return value == comparator
	default:
		log.Println("cannot perform an equals operation on this type")
		return false
	}
}
