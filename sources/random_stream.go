package main

import (
	"flag"
)

var (
	rate = flag.Float64("rate", "1", "Poisson rate at which to generate messages")
	topic = flag.String("write_topic", "", "topic to write to")
)
