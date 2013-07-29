package main

import (
	"flag"
)

var (
	readTopic = flag.String("read_topic", "", "topic to read from")
	writeTopic = flag.String("write_topic", "", "topic to write to")
	key = flag.String("reduce_key", "", "key against which to reduce")
	timeout = flag.Float64("reduce_timeout", 0, "how long to wait before emitting")
    maxLength = flag.Int("max_length", 0, "maximum length of an emitted message")
)
