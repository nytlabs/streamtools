package streamtools

import (
	"github.com/bitly/go-simplejson"
	"log"
)

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
