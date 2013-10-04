package main

import (
    "flag"
    "github.com/nytlabs/streamtools/streamtools"
)

var (
    topic    = flag.String("topic", "", "topic to read from")
    rulePort = flag.String("port", "8080", "port to listen for new rules on")
    name     = flag.String("name", "last_state", "name of block")
)

func main() {
    flag.Parse()
    streamtools.SetupLogger(name)
    block := streamtools.NewStateBlock(streamtools.LastState, "laststate")
    block.Run(*topic, *rulePort)
}
