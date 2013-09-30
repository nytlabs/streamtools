package main

import (
	"flag"
	"github.com/nytlabs/streamtools/streamtools"
)

var (
	topic = flag.String("topic", "", "topic to read from")
	name  = flag.String("name", "to-stdout", "name of block")
)

func main() {
	flag.Parse()
	block := streamtools.NewInBlock(streamtools.ToStdout, *name)
	block.Run(*topic, "8080")
}
