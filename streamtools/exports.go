package streamtools

import (
	"github.com/bitly/go-simplejson"
	"log"
)

// StreamToStdOut simply prints the JSON string of each message.
var StreamToStdOut ExportFunction = func(inChan chan simplejson.Json) {
	for {
		select {
		case m := <-inChan:
			blob, err := m.MarshalJSON()
			if err != nil {
				log.Fatalln(err)
			}
			log.Println(string(blob))
		}
	}
}
