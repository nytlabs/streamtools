package main

import (
	"flag"
)

var (
	url = flag.String("url", "", "streaming url endpoint")
        inboundTopic = flag.String("read_topic", "", "topic to read from")	
)
