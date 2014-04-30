package tests

import (
	"log"
	"reflect"
	"time"

	"github.com/nytlabs/streamtools/st/blocks"
	"github.com/nytlabs/streamtools/st/loghub"
	"github.com/nytlabs/streamtools/test_utils"
	. "launchpad.net/gocheck"
)

type RedisSuite struct{}

var redisSuite = Suite(&RedisSuite{})

func (s *RedisSuite) TestRedisSADD(c *C) {
	loghub.Start()
	log.Println("testing Redis: SADD")

	b, ch := test_utils.NewBlock("testingRedisSADD", "redis")
	go blocks.BlockRoutine(b)

	m := []string{"'foobar'", "'baz'"}
	ruleMsg := map[string]interface{}{"Server": "localhost:6379", "Password": "", "Command": "SADD", "Arguments": m}
	toRule := &blocks.Msg{Msg: ruleMsg, Route: "rule"}
	ch.InChan <- toRule

	outChan := make(chan *blocks.Msg)
	ch.AddChan <- &blocks.AddChanMsg{Route: "1", Channel: outChan}

	ch.InChan <- &blocks.Msg{Msg: map[string]interface{}{"tick": "123"}, Route: "in"}

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

		case messageI := <-outChan:
			message := messageI.Msg.(map[string]interface{})
			c.Assert(message["response"], FitsTypeOf, float64(0))

		case err := <-ch.ErrChan:
			if err != nil {
				c.Errorf(err.Error())
			} else {
				return
			}
		}
	}
}

func (s *RedisSuite) TestRedisSMEMBERS(c *C) {
	loghub.Start()
	log.Println("testing Redis: SMEMBERS")

	b, ch := test_utils.NewBlock("testingRedisSMEM", "redis")
	go blocks.BlockRoutine(b)

	m := []string{"'foobar'"}
	ruleMsg := map[string]interface{}{"Server": "localhost:6379", "Password": "", "Command": "SMEMBERS", "Arguments": m}
	toRule := &blocks.Msg{Msg: ruleMsg, Route: "rule"}
	ch.InChan <- toRule

	outChan := make(chan *blocks.Msg)
	ch.AddChan <- &blocks.AddChanMsg{Route: "1", Channel: outChan}

	ch.InChan <- &blocks.Msg{Msg: map[string]interface{}{"tick": "123"}, Route: "in"}

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

		case messageI := <-outChan:
			message := messageI.Msg.(map[string]interface{})
			expectedValue := []interface{}{"baz"}
			c.Assert(message["response"], DeepEquals, expectedValue)

		case err := <-ch.ErrChan:
			if err != nil {
				c.Errorf(err.Error())
			} else {
				return
			}
		}
	}
}

func (s *RedisSuite) TestRedisDBSIZE(c *C) {
	loghub.Start()
	log.Println("testing Redis: DBSIZE")

	b, ch := test_utils.NewBlock("testingRedisDBSIZE", "redis")
	go blocks.BlockRoutine(b)

	var args = make([]string, 0, 10)
	ruleMsg := map[string]interface{}{"Server": "localhost:6379", "Password": "", "Command": "DBSIZE", "Arguments": args}
	toRule := &blocks.Msg{Msg: ruleMsg, Route: "rule"}
	ch.InChan <- toRule

	outChan := make(chan *blocks.Msg)
	ch.AddChan <- &blocks.AddChanMsg{Route: "1", Channel: outChan}

	ch.InChan <- &blocks.Msg{Msg: map[string]interface{}{"tick": "123"}, Route: "in"}

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

		case messageI := <-outChan:
			message := messageI.Msg.(map[string]interface{})
			c.Assert(message["response"], FitsTypeOf, float64(0))

		case err := <-ch.ErrChan:
			if err != nil {
				c.Errorf(err.Error())
			} else {
				return
			}
		}
	}
}
