// This binary reads from an nsq stream, and writes to another stream
// the corresponding messages with flattened keys and JSON type.
package main

import (
	"flag"
	"github.com/nytlabs/streamtools/streamtools"
)

var (
	inTopic          = flag.String("in_topic", "", "topic to read from")
	outTopic         = flag.String("out_topic", "", "topic to write to")
	channel          = flag.String("channel", "type_inferencer", "nsq reader name")
)

func main() {
	flag.Parse()
	// var InferType streamtools.STFunc = streamtools.InferType
	streamtools.TransferBlock(*inTopic, *outTopic, *channel, streamtools.InferType)
}