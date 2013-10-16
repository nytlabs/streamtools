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
	log.Println("building library")
	library = make(map[string]BlockTemplate)
	library["ticker"] = BlockTemplate{
		blockType:        "ticker",
		blockFactory:     NewTicker,
		blockDescription: "outputs the time every n seconds",
	}
}
