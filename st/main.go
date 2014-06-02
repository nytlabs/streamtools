package main

import (
	"flag"
	"github.com/nytlabs/streamtools/st/library"
	"github.com/nytlabs/streamtools/st/loghub"
	"github.com/nytlabs/streamtools/st/server"
	"github.com/nytlabs/streamtools/st/util"
	"log"
	"os"
)

var (
	// port that streamtools reuns on
	port    = flag.String("port", "7070", "streamtools port")
	domain  = flag.String("domain", "127.0.0.1", "streamtools domain")
	version = flag.Bool("version", false, "prints current streamtools version")
)

func main() {
	flag.Parse()

	if *version {
		log.Println("Streamtools version:", util.VERSION)
		os.Exit(0)
	}

	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	library.Start()
	loghub.Start()

	s := server.NewServer()

	for _, file := range flag.Args() {
		s.ImportFile(file)
	}

	s.Id = "SERVER"
	s.Port = *port
	s.Domain = *domain

	s.Run()
}
