package main

import (
	"flag"
)

var (
	readTopic  = flag.String("read_topic", "", "topic to read from")
	writeTopic = flag.String("write_topic", "", "topic to write to")
)
