package main

import (
    "flag"
	"github.com/nytlabs/streamtools/daemon"
)

var (
    // port that streamtools reuns on
    port = flag.String("port", "7070", "stream tools port")
)


func main() {
    flag.Parse()
    d := daemon.Daemon{}
    d.Run(*port)
}
