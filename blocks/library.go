package blocks

import (
	"log"
)

var (
	Library     BlockLibrary
	libraryBlob string
)

// A block library is a collection of possible block templates
type BlockLibrary map[string]Block

func BuildLibrary() {
	log.Println("building block library")
	Library = make(map[string]Block)

	// BLOCKS

	Library["ticker"] = Block{
		Template: &BlockTemplate{
			BlockType:  "ticker",
			RouteNames: []string{},
		},
		Routine: Ticker,
	}

	Library["connection"] = Block{
		Template: &BlockTemplate{
			BlockType:  "connection",
			RouteNames: []string{"last_seen"},
		},
		Routine: Connection,
	}

	Library["tolog"] = Block{
		Template: &BlockTemplate{
			BlockType:  "tolog",
			RouteNames: []string{},
		},
		Routine: ToLog,
	}
}
