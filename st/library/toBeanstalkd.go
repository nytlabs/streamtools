package library

import (
	"encoding/json"
	"errors"

	"github.com/nutrun/lentil"
	"github.com/nytlabs/streamtools/st/blocks" // blocks
	"github.com/nytlabs/streamtools/st/util"
)

// specify those channels we're going to use to communicate with streamtools
type ToBeanstalkd struct {
	blocks.Block
	queryrule chan blocks.MsgChan
	inrule    blocks.MsgChan
	in        blocks.MsgChan
	quit      blocks.MsgChan
}

// we need to build a simple factory so that streamtools can make new blocks of this kind
func NewToBeanstalkd() blocks.BlockInterface {
	return &ToBeanstalkd{}
}

// Setup is called once before running the block. We build up the channels and specify what kind of block this is.
func (b *ToBeanstalkd) Setup() {
	b.Kind = "Queues"
	b.Desc = "sends jobs to beanstalkd tube"
	b.in = b.InRoute("in")
	b.inrule = b.InRoute("rule")
	b.queryrule = b.QueryRoute("rule")
	b.quit = b.Quit()
}

// Run is the block's main loop. Here we listen on the different channels we set up.
func (b *ToBeanstalkd) Run() {
	var conn *lentil.Beanstalkd
	var tube = "default"
	var ttr = 0
	var host = ""
	var err error
	for {
		select {
		case msgI := <-b.inrule:
			// set hostname for beanstalkd server
			host, err = util.ParseString(msgI, "Host")
			if err != nil {
				b.Error(err.Error())
				continue
			}
			// set tube name
			tube, err = util.ParseString(msgI, "Tube")
			if err != nil {
				b.Error(errors.New("Could not parse tube name, setting to 'default'"))
				tube = "default"
			}
			// set time to reserve
			ttr, err = util.ParseInt(msgI, "TTR")
			if err != nil || ttr < 0 {
				b.Error(errors.New("Error parsing TTR. Setting TTR to 0"))
				ttr = 0
			}
			// create beanstalkd connection
			conn, err = lentil.Dial(host)
			if err != nil {
				// swallowing a panic from lentil here - streamtools must not die
				b.Error(errors.New("Could not initiate connection with beanstalkd server"))
				continue
			}
			// use the specified tube
			conn.Use(tube)
		case <-b.quit:
			// close connection to beanstalkd and quit
			if conn != nil {
				conn.Quit()
			}
			return
		case msg := <-b.in:
			// deal with inbound data
			msgStr, err := json.Marshal(msg)
			if err != nil {
				b.Error(err)
				continue
			}
			if conn != nil {
				_, err := conn.Put(0, 0, ttr, msgStr)
				if err != nil {
					b.Error(err.Error())
				}
			} else {
				b.Error(errors.New("Beanstalkd connection not initated or lost. Please check your beanstalkd server or block settings."))
			}
		case MsgChan := <-b.queryrule:
			// deal with a query request
			MsgChan <- map[string]interface{}{
				"Host": host,
				"Tube": tube,
				"TTR":  ttr,
			}
		}
	}
}
