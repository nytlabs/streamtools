package library

import (
	"encoding/json"
	"github.com/nikhan/go-sqsReader"           //sqsReader
	"github.com/nytlabs/streamtools/st/blocks" // blocks
	"github.com/nytlabs/streamtools/st/util"
)

// specify those channels we're going to use to communicate with streamtools
type FromSQS struct {
	blocks.Block
	queryrule    chan chan interface{}
	inrule       chan interface{}
	out          chan interface{}
	fromReader   chan []byte
	quit         chan interface{}
	SQSEndpoint  string
	AccessKey    string
	AccessSecret string
}

// we need to build a simple factory so that streamtools can make new blocks of this kind
func NewFromSQS() blocks.BlockInterface {
	return &FromSQS{}
}

// Setup is called once before running the block. We build up the channels and specify what kind of block this is.
func (b *FromSQS) Setup() {
	b.Kind = "fromSQS"
	b.inrule = b.InRoute("rule")
	b.queryrule = b.QueryRoute("rule")
	b.quit = b.Quit()
	b.out = b.Broadcast()
}

// Run is the block's main loop. Here we listen on the different channels we set up.
func (b *FromSQS) Run() {
	var SQSEndpoint, AccessKey, AccessSecret string
	var err error
	fromReader := make(chan []byte)
	for {
		select {
		case msgI := <-b.inrule:
			SQSEndpoint, err = util.ParseString(msgI, "SQSEndpoint")
			if err != nil {
				b.Error(err)
				break
			}

			AccessKey, err = util.ParseString(msgI, "AccessKey")
			if err != nil {
				b.Error(err)
				break
			}

			AccessSecret, err = util.ParseString(msgI, "AccessSecret")
			if err != nil {
				b.Error(err)
				break
			}

			r := sqsReader.NewReader(SQSEndpoint, AccessKey, AccessSecret, fromReader)
			go r.Start()

		case msg := <-fromReader:
			var outMsg interface{}
			err := json.Unmarshal(msg, &outMsg)
			if err != nil {
				b.Error(err)
				continue
			}
			b.out <- outMsg
		case <-b.quit:
			// quit the block
			return
		case respChan := <-b.queryrule:
			// deal with a query request
			respChan <- map[string]interface{}{
				"SQSEndpoint":  SQSEndpoint,
				"AccessKey":    AccessKey,
				"AccessSecret": AccessSecret,
			}
		}
	}
}
