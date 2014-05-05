package tests

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"reflect"
	"time"

	"github.com/nytlabs/streamtools/st/blocks"
	"github.com/nytlabs/streamtools/st/loghub"
	"github.com/nytlabs/streamtools/test_utils"
	. "launchpad.net/gocheck"
)

type FromSQSSuite struct{}

var fromSQSSuite = Suite(&FromSQSSuite{})

func (s *FromSQSSuite) TestFromSQS(c *C) {
	loghub.Start()
	log.Println("testing FromSQS")

	sampleResponse := string(`
<CreateQueueResponse
  xmlns="http://sqs.us-east-1.amazonaws.com/doc/2012-11-05/"
  xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
  xsi:type="CreateQueueResponse">
     <CreateQueueResult>
        <QueueUrl>
        http://sqs.us-east-1.amazonaws.com/770098461991/queue2
        </QueueUrl>
     </CreateQueueResult>
     <ResponseMetadata>
        <RequestId>cb919c0a-9bce-4afe-9b48-9bdf2412bb67</RequestId>
     </ResponseMetadata>
</CreateQueueResponse>
  `)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, sampleResponse)
	}))
	defer ts.Close()

	b, ch := test_utils.NewBlock("testingFromSQS", "fromsqs")
	go blocks.BlockRoutine(b)

	ruleMsg := map[string]interface{}{"SQSEndpoint": ts.URL, "AccessKey": "123access", "AccessSecret": "123secret", "APIVersion": "2012-11-05", "SignatureVersion": "4", "WaitTimeSeconds": "0", "MaxNumberOfMessages": "10"}
	toRule := &blocks.Msg{Msg: ruleMsg, Route: "rule"}
	ch.InChan <- toRule

	outChan := make(chan *blocks.Msg)
	ch.AddChan <- &blocks.AddChanMsg{Route: "1", Channel: outChan}

	queryOutChan := make(blocks.MsgChan)
	ch.QueryChan <- &blocks.QueryMsg{MsgChan: queryOutChan, Route: "rule"}

	time.AfterFunc(time.Duration(5)*time.Second, func() {
		ch.QuitChan <- true
	})

	for {
		select {
		case messageI := <-queryOutChan:
			if !reflect.DeepEqual(messageI, ruleMsg) {
				log.Println("Rule mismatch:", messageI, ruleMsg)
				c.Fail()
			}

		case message := <-outChan:
			log.Println(message)

		case err := <-ch.ErrChan:
			if err != nil {
				c.Errorf(err.Error())
			} else {
				return
			}
		}
	}
}
