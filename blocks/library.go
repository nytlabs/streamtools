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

// TODO: library should probably become a struct at some point...
func AddBlock(name string, routeNames []string, routine BlockRoutine){
	Library[name] = Block{
		Template: &BlockTemplate{
			BlockType: name,
			RouteNames: routeNames,
		},
		Routine: routine,
	}
}

func BuildLibrary() {
	log.Println("building block library")
	Library = make(map[string]Block)

	// BLOCKS
	AddBlock("ticker", []string{}, Ticker)
	AddBlock("connection", []string{"last_seen"}, Connection)
	AddBlock("tolog", []string{}, ToLog)
}
