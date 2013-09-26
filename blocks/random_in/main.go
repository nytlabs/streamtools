package main

import (
	"flag"
	"github.com/nytlabs/stream_tools/streamtools"
)

var (
	topic = flag.String("topic", "", "topic to write to")
)

func main() {
	flag.Parse()
	random := streamtools.NewOutBlock(streamtools.Random, "random_stream")
	random.Run(*topic, "8081")
}
