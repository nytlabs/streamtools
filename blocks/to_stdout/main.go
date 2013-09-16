package main

import (
	"flag"
	"github.com/mikedewar/stream_tools/streamtools"
)

var (
	topic = flag.String("topic", "", "topic to read from")
)

func main() {
	flag.Parse()
	to_stdout := streamtools.NewInBlock(streamtools.ToStdout, "to_stdout")
	to_stdout.Run(*topic, "8080")
}
