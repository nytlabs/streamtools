package main

import (
	"flag"
	"github.com/nytlabs/streamtools/streamtools"
	"log"
	"os"
)

var (
	topic = flag.String("topic", "", "topic to write to")
	name  = flag.String("name", "random_in", "name of block")
)

func main() {
	flag.Parse()

	log.SetPrefix(" [" + *name + "] ")
	logfile, err := os.Create(*name + ".log")
	if err != nil {
		log.Fatal(err.Error())
	}
	log.SetOutput(logfile)

	random := streamtools.NewOutBlock(streamtools.Random, *name)
	random.Run(*topic, "8081")
}
