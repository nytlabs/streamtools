package streamtools

import (
	"github.com/bitly/go-simplejson"
	"log"
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
	log.Println("broadcasting on", channels)
	for name, c := range channels {
		log.Println("broadcasting to", name)
		c <- msg
		log.Println("broadcast complete")
	}
}
