package main

import (
	"flag"
	"github.com/nytlabs/streamtools/st/server"
	"github.com/nytlabs/streamtools/st/library"
	"github.com/nytlabs/streamtools/st/loghub"
	"log"
)

var (
	// port that streamtools reuns on
	port   = flag.String("port", "7070", "streamtools port")
	domain = flag.String("domain", "127.0.0.1", "streamtools domain")
)

func main() {
	flag.Parse()
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	library.Start()
	loghub.Start()

	s := server.NewServer()

	s.Id = "SERVER"
	s.Port = *port
	s.Domain = *domain

	s.Run()
}
