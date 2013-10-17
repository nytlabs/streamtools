package streamtools

import (
	"github.com/bitly/go-simplejson"
	"strconv"
)

func IDService(idChan chan string) {
	i := 1
	for {
		id := strconv.Itoa(i)
		idChan <- id
		i += 1
	}
}

func broadcast(channels map[string]chan *simplejson.Json, msg *simplejson.Json) {
	for _, c := range channels {
		c <- msg
	}
}
