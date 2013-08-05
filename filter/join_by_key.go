package main

import (
	"flag"
)

var (
	key      = flag.String("key", "", "key to join on")
	primaryTopic  = flag.String("primary_topic", "", "primary topic to read from")
	secondaryTopic  = flag.String("secondary_topic", "", "secondary topic to read from")
	outTopic = flag.String("out_topic", "", "topic to write to")
    timeout = flag.String("timoue", "", "amount of time to wait for a matching message before giving up")
)
