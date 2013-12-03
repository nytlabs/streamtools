package blocks

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
)

func ToFile(b *Block) {

	type toFileRule struct {
		Filename string
	}

	var rule *toFileRule
	var w *bufio.Writer

	for {
		select {
		case msg := <-b.InChan:
			if rule == nil {
				break
			}

			msgStr, err := json.Marshal(msg)
			if err != nil {
				log.Println("wow bad json")
			}

			fmt.Fprintln(w, string(msgStr))
			w.Flush()

		case msg := <-b.Routes["set_rule"]:
			if rule == nil {
				rule = &toFileRule{}
			}

			unmarshal(msg, rule)

			fo, err := os.Create(rule.Filename)
			if err != nil {
				log.Fatal(err.Error())
			}
			defer fo.Close()
			w = bufio.NewWriter(fo)

		case msg := <-b.Routes["get_rule"]:
			if rule == nil {
				marshal(msg, &toFileRule{})
			} else {
				marshal(msg, rule)
			}
		case <-b.QuitChan:
			quit(b)
			return
		}
	}
}
