package blocks

import (
	"log"
)

var (
	Library     BlockLibrary
	libraryBlob string
)

// A block library is a collection of possible block templates
type BlockLibrary map[string]*BlockTemplate

func BuildLibrary() {
	log.Println("building block library")
	Library = make(map[string]*BlockTemplate)

	templates := []*BlockTemplate{
		&BlockTemplate{
			BlockType: "ticker",
			RouteNames: []string{},
			Routine: Ticker,
		},
		&BlockTemplate{
			BlockType: "connection",
			RouteNames: []string{"last_seen"},
			Routine: Connection,
		},
		&BlockTemplate{
			BlockType: "tolog",
			RouteNames: []string{},
			Routine: ToLog,
		},
		&BlockTemplate{
			BlockType: "random",
			RouteNames: []string{},
			Routine: Random,
		},
		&BlockTemplate{
			BlockType: "count",
			RouteNames: []string{"params", "count"},
			Routine: Count,
		},
	}

	for _, t := range templates {
		Library[t.BlockType] = t
	}
}
