package main

import (
	"flag"
)

var (
	key      = flag.String("key", "", "key to filter by")
	inTopic  = flag.String("in_topic", "", "topic to read from")
	outTopic = flag.String("out_topic", "", "topic to write to")
)
