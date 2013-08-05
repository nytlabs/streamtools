package main

import (
	"flag"
)


var (
	filename = flag.String("filename", "", "file name of the csv file")
	topic = flag.String("write_topic", "", "topic to write to")
)
