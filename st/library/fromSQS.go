package library

import (
	"encoding/json"
	"github.com/nikhan/go-sqsReader"           //sqsReader
	"github.com/nytlabs/streamtools/st/blocks" // blocks
	"github.com/nytlabs/streamtools/st/util"
	"log"
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
		log.Println("fromSQS: for")
		select {
		case msgI := <-b.inrule:
			log.Println("fromSQS: inrule")
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

			log.Println(SQSEndpoint)
			log.Println(AccessKey)
			log.Println(AccessSecret)
			r := sqsReader.NewReader(SQSEndpoint, AccessKey, AccessSecret, fromReader)
			log.Println("fromSQS starting reader")
			go r.Start()
			log.Println("fromSQS started reader")

		case msg := <-fromReader:
			log.Println("fromSQS: fromReader")
			var outMsg interface{}
			err := json.Unmarshal(msg, &outMsg)
			if err != nil {
				b.Error(err)
				continue
			}
			log.Println("fromSQS: sending Out")
			b.out <- outMsg
			log.Println("fromSQS: sent Out")
		case <-b.quit:
			log.Println("fromSQS: quit")
			// quit the block
			return
		case respChan := <-b.queryrule:
			log.Println("fromSQS: query rule")
			// deal with a query request
			respChan <- map[string]interface{}{
				"SQSEndpoint":  SQSEndpoint,
				"AccessKey":    AccessKey,
				"AccessSecret": AccessSecret,
			}
			log.Println("fromSQS: sent response")
		}
	}
}
