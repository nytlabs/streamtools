package main

import (
	"flag"
	"github.com/nytlabs/streamtools/daemon"
	"log"
)

var (
	// port that streamtools reuns on
	port   = flag.String("port", "7070", "streamtools port")
	domain = flag.String("domain", "127.0.0.1","streamtools domain")
)

func main() {
	flag.Parse()
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	d := daemon.NewDaemon()

	d.Port = *port
	d.Domain = *domain

	d.Run()
}
