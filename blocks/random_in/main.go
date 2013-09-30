package main

import (
	"flag"
	"github.com/nytlabs/streamtools/streamtools"
)

var (
	topic = flag.String("topic", "", "topic to write to")
	name  = flag.String("name", "random_in", "name of block")
)

func main() {
	flag.Parse()

	streamtools.SetupLogger(name)

	random := streamtools.NewOutBlock(streamtools.Random, *name)
	random.Run(*topic, "8081")
}
