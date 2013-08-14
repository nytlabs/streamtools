package streamtools

import (
	"fmt"
	"github.com/bitly/go-simplejson"
	"log"
	"os"
)

func ToStdout(inChan chan simplejson.Json, ruleChan chan *simplejson.Json) {
	for {
		select {
		case <-ruleChan:
		case msg := <-inChan:
			out, err := msg.Encode()
			if err != nil {
				log.Fatal(err.Error())
			}
			_, err = fmt.Println(os.Stdout, string(out))
			if err != nil {
				log.Fatal(err.Error())
			}
		}
	}

}
