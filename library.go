package streamtools

import (
	"log"
)

var (
	library BlockLibrary
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
		blockFactory:	  NewToLog,
		blockDescription: "prints messages in log",
	}


}
