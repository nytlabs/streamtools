package streamtools

import (
	"github.com/bitly/go-simplejson"
	"io/ioutil"
	"log"
	"os/exec"
)

func JQ(inChan chan *simplejson.Json, outChan chan *simplejson.Json, RuleChan chan *simplejson.Json) {

	rule := <-RuleChan
	command, err := rule.Get("command").String()
	if err != nil {
		log.Println(rule)
		log.Fatal(err.Error())
	}

	for {
		select {
		case msg := <-inChan:

			cmd := exec.Command("jq", command)
			stdin, err := cmd.StdinPipe()
			if err != nil {
				log.Fatal(err.Error())
			}
			stdout, err := cmd.StdoutPipe()
			if err != nil {
				log.Fatal(err.Error())
			}
			err = cmd.Start()
			if err != nil {
				log.Fatal(err.Error())
			}
			inBytes, err := msg.Encode()
			if err != nil {
				log.Fatal(err.Error())
			}
			stdin.Write(inBytes)
			outBytes, err := ioutil.ReadAll(stdout)
			if err != nil {
				log.Fatal(err.Error())
			}
			outMsg, err := simplejson.NewJson(outBytes)
			if err != nil {
				log.Fatal(err.Error())
			}
			outChan <- outMsg

		}
	}

}
