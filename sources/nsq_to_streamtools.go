package main

import (
	"flag"
)

var (
	inboundTopic = flag.String("read_topic", "", "topic to read from")
	streamtoolsTopic = flag.String("write_topic", "", "topic to write to")
)
