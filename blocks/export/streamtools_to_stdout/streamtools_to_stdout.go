package main

import (
	"flag"
	"github.com/nytlabs/streamtools/streamtools"
)

var (
	inTopic = flag.String("in_topic", "", "topic to read from")
	channel = flag.String("channel", "stdout", "nsq reader name")
)

func main() {
	flag.Parse()
	streamtools.ExportBlock(*inTopic, *channel, streamtools.StreamToStdOut)
}
