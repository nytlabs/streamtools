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

	rule := &toFileRule{}

	unmarshal(<-b.Routes["set_rule"], &rule)

	fo, err := os.Create(rule.Filename)
	if err != nil {
		log.Fatal(err.Error())
	}
	defer fo.Close()
	w := bufio.NewWriter(fo)

	for {
		select {
		case msg := <-b.InChan:
			msgStr, err := json.Marshal(msg)
			if err != nil {
				log.Println("wow bad json")
			}
			fmt.Fprintln(w, string(msgStr))
			w.Flush()
		case msg := <-b.AddChan:
			updateOutChans(msg, b)
		case <-b.QuitChan:
			quit(b)
			return
		}
	}
}
