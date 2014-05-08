package tests

import (
	"io/ioutil"
	"log"
	"syscall"
	//	"syscall"
	"time"

	"github.com/nytlabs/streamtools/st/blocks"
	"github.com/nytlabs/streamtools/test_utils"
	. "launchpad.net/gocheck"
)

type FromFileSuite struct{}

var fromFileSuite = Suite(&FromFileSuite{})

func (s *FromFileSuite) TestFromFile(c *C) {
	log.Println("testing FromFile")
	b, ch := test_utils.NewBlock("testingFile", "fromfile")
	go blocks.BlockRoutine(b)
	outChan := make(chan *blocks.Msg)
	ch.AddChan <- &blocks.AddChanMsg{
		Route:   "out",
		Channel: outChan,
	}

	f, err := ioutil.TempFile("", "streamtools_test_from_file.log")
	if err != nil {
		c.Errorf(err.Error())
	}

	defer syscall.Unlink(f.Name())

	var fromfilestring = string(`{"Name": "Jacqui Maher", "Location": "Brooklyn", "Dog": "Conor S. Dogberst" }
{"Name": "Nik Hanselmann", "Location": "New York", "Dog": "None:(" }
{"Name": "Mike Dewar", "Location": "The Moon", "Dog": "Percy ? Dewar" }`)

	ioutil.WriteFile(f.Name(), []byte(fromfilestring), 0644)

	ruleMsg := map[string]interface{}{"Filename": f.Name()}
	toRule := &blocks.Msg{Msg: ruleMsg, Route: "rule"}
	ch.InChan <- toRule

	time.AfterFunc(time.Duration(1)*time.Second, func() {
		ch.InChan <- &blocks.Msg{Msg: map[string]interface{}{}, Route: "poll"}
	})
	time.AfterFunc(time.Duration(1)*time.Second, func() {
		ch.InChan <- &blocks.Msg{Msg: map[string]interface{}{}, Route: "poll"}
	})
	time.AfterFunc(time.Duration(1)*time.Second, func() {
		ch.InChan <- &blocks.Msg{Msg: map[string]interface{}{}, Route: "poll"}
	})

	time.AfterFunc(time.Duration(5)*time.Second, func() {
		ch.QuitChan <- true
	})

	var expectedNames = []string{"Jacqui Maher", "Nik Hanselmann", "Mike Dewar"}
	var expectedLocations = []string{"Brooklyn", "New York", "The Moon"}
	for {
		select {
		case err := <-ch.ErrChan:
			if err != nil {
				c.Errorf(err.Error())
			} else {
				return
			}
		case messageI := <-outChan:
			message := messageI.Msg.(map[string]interface{})

			nameReceived, ok := message["Name"].(string)
			if !ok {
				log.Println("failed asserting message['Name'] to a string")
			}

			locationReceived, ok := message["Location"].(string)
			if !ok {
				log.Println("failed asserting message['Location'] to a string")
			}

			if !test_utils.StringInSlice(expectedNames, nameReceived) {
				log.Println("failed finding", nameReceived, "in expected names list")
				c.Fail()
			}

			if !test_utils.StringInSlice(expectedLocations, locationReceived) {
				log.Println("failed finding", locationReceived, "in expected locations list")
				c.Fail()
			}
		}
	}
}
