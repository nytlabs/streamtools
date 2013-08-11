// This binary reads from an nsq stream, and writes to another stream
// describing the difference between consecutive messages; this transfer
// block outputs n-1 messages for every n input messages.
package main

import (
	"flag"
	"github.com/nytlabs/streamtools/streamtools"
)

var (
	inTopic  = flag.String("in_topic", "", "topic to read from")
	outTopic = flag.String("out_topic", "", "topic to write to")
	channel  = flag.String("channel", "diff_transform", "nsq reader name")
)

func main() {
	flag.Parse()
	streamtools.TransferBlock(*inTopic, *outTopic, *channel, streamtools.Diff)
}
