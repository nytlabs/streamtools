package main

import (
	"flag"
	"github.com/nytlabs/streamtools/daemon"
	"log"
)

var (
	// port that streamtools reuns on
	port   = flag.String("port", "7070", "stream tools port")
	config = flag.String("config", "", "config .json")
)

func main() {
	flag.Parse()
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	d := daemon.Daemon{
		Port:   *port,
		Config: *config,
	}

	d.Run()
}
