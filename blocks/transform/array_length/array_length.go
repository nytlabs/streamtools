// This binary reads from an nsq stream, and writes to another stream 
// the corresponding messages with length of array at specified key.
package main

import (
	"flag"
	"github.com/nytlabs/streamtools/streamtools"
	"github.com/bitly/go-simplejson"
	"github.com/bitly/nsq/nsq"
)

var (
	inTopic          = flag.String("in_topic", "", "topic to read from")
	outTopic         = flag.String("out_topic", "", "topic to write to")
	channel          = flag.String("channel", "array_length", "nsq reader name")
	arrayKey          = flag.String("arrayKey", "", "obj key whose length")
)

func main() {
	flag.Parse()
	f := func(msgChan chan *nsq.Message, outChan chan simplejson.Json) {
			streamtools.GetArrayLength(*arrayKey, msgChan, outChan)
	}
	streamtools.TransferBlock( *inTopic, *outTopic, *channel, f )
}