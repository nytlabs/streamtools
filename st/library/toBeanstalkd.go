package library

import (
	"encoding/json"
	"github.com/nytlabs/streamtools/st/blocks" // blocks
	"github.com/nytlabs/streamtools/st/util"
    "github.com/nutrun/lentil"
    "errors"
)

// specify those channels we're going to use to communicate with streamtools
type ToBeanstalkd struct {
	blocks.Block
	host      string
    tube      string
    ttr       int
	queryrule chan chan interface{}
	inrule    chan interface{}
	in        chan interface{}
	quit      chan interface{}
}

// we need to build a simple factory so that streamtools can make new blocks of this kind
func NewToBeanstalkd() blocks.BlockInterface {
	return &ToBeanstalkd{}
}

// Setup is called once before running the block. We build up the channels and specify what kind of block this is.
func (b *ToBeanstalkd) Setup() {
	b.Kind = "ToBeanstalkd"
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
	for {
		select {
		case msgI := <-b.inrule:
			// set hostname for beanstalkd server
			host, err := util.ParseString(msgI, "Host")
            if err != nil {
                b.Error(err.Error())
                continue
            } 
            b.host = host
            // create beanstalkd connection
			conn, err = lentil.Dial(b.host)
            if err != nil {
                // swallowing a panic from lentil here - streamtools must not die
                b.Error(errors.New("Could not initiate connection with beanstalkd server"))
                continue
            }
            // set tube name
            tube, _ = util.ParseString(msgI, "Tube")
            b.tube = tube
            // use the specified tube
            conn.Use(b.tube)
            // set time to reserve
            ttr, err = util.ParseInt(msgI, "TTR")
            if err !=nil || ttr < 0 {
                //b.Error(errors.New("Error parsing TTR. Setting TTR to 0"))
                b.Error(err.Error())
                ttr = 0
            }
            b.ttr = ttr
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
			}
            if conn != nil {
                _, err :=  conn.Put(0, 0 , ttr, msgStr)
                if err != nil {
                    b.Error(err.Error())
                }
            } else {
                b.Error(errors.New("Beanstalkd connection not initated or lost. Please check your beanstalkd server or block settings."))
            }
		case respChan := <-b.queryrule:
			// deal with a query request
			respChan <- map[string]interface{}{
				"Host": b.host,
                "Tube": b.tube,
                "TTR" : b.ttr,
			}
		}
	}
}
