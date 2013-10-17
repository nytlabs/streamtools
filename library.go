package streamtools

import (
	"encoding/json"
	"log"
)

var (
	library     BlockLibrary
	libraryBlob string
)

// A block template is a definition of a sprcific block
type BlockTemplate struct {
	blockType        string
	blockFactory     func() Block
	blockDescription string
}

// A block library is a collection of possible block templates
type BlockLibrary map[string]BlockTemplate

func buildLibrary() {
	log.Println("building block library")
	library = make(map[string]BlockTemplate)

	// BLOCKS
	library["connection"] = BlockTemplate{
		blockType:        "connection",
		blockFactory:     NewConnection,
		blockDescription: "connects to blocks",
	}

	library["ticker"] = BlockTemplate{
		blockType:        "ticker",
		blockFactory:     NewTicker,
		blockDescription: "outputs the time every n seconds",
	}

	library["lastseen"] = BlockTemplate{
		blockType:        "lastseen",
		blockFactory:     NewLastSeen,
		blockDescription: "poll block tells you what the last message is",
	}

	library["tolog"] = BlockTemplate{
		blockType:        "tolog",
		blockFactory:     NewToLog,
		blockDescription: "prints messages in log",
	}

	library["routeexample"] = BlockTemplate{
		blockType:        "routeexample",
		blockFactory:     NewRouteExample,
		blockDescription: "an example for routing.",
	}

	// create a json blob that contains the library block list
	blockList := make([]map[string]string, len(library))
	i := 0
	for k, v := range library {
		blockMeta := make(map[string]string)
		blockMeta[k] = v.blockDescription
		blockList[i] = blockMeta
		i++
	}
	lj, _ := json.Marshal(blockList)
	libraryBlob = string(lj)
}
