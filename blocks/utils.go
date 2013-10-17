package blocks

import (
	"github.com/bitly/go-simplejson"
)

func broadcast(channels map[string]chan *simplejson.Json, msg *simplejson.Json) {
	for _, c := range channels {
		c <- msg
	}
}
