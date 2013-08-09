// This binary reads from an nsq stream, and writes to a store which
// serves it contents as JSON on HTTP.
package main

import (
	"flag"
	"github.com/nytlabs/streamtools/streamtools"
)

var (
	inTopic = flag.String("in_topic", "", "topic to read from")
	route   = flag.String("route", "/", "localhost address on which to serve http")
	port    = flag.Int("port", 8080, "localhost port on which to serve http")
	channel = flag.String("channel", "max", "nsq reader name")
)

func main() {
	flag.Parse()
	streamtools.TrackingBlock(*inTopic, *channel, *route, *port, streamtools.TrackMax)
}
