package main

import (
	"flag"
)

var (
	wsPort       = flag.String("port", "", "websocket port")
	inboundTopic = flag.String("read_topic", "", "topic to read from")
)
