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
			BlockType:  "ticker",
			RouteNames: []string{"set_rule", "get_rule"},
			Routine:    Ticker,
		},
		&BlockTemplate{
			BlockType:  "connection",
			RouteNames: []string{"last_seen"},
			Routine:    Connection,
		},
		&BlockTemplate{
			BlockType:  "tolog",
			RouteNames: []string{},
			Routine:    ToLog,
		},
		&BlockTemplate{
			BlockType:  "random",
			RouteNames: []string{"set_rule"},
			Routine:    Random,
		},
		&BlockTemplate{
			BlockType:  "count",
			RouteNames: []string{"set_rule", "count"},
			Routine:    Count,
		},
		&BlockTemplate{
			BlockType:  "mask",
			RouteNames: []string{"set_rule", "get_rule"},
			Routine:    Mask,
		},
		&BlockTemplate{
			BlockType:  "sync",
			RouteNames: []string{"set_rule", "get_rule"},
			Routine:    Sync,
		},
		&BlockTemplate{
			BlockType:  "post",
			RouteNames: []string{"set_rule"},
			Routine:    Post,
		},
		&BlockTemplate{
			BlockType:  "date",
			RouteNames: []string{"set_rule"},
			Routine:    Date,
		},
		&BlockTemplate{
			BlockType:  "fromNSQ",
			RouteNames: []string{"set_rule"},
			Routine:    FromNSQ,
		},
		&BlockTemplate{
			BlockType:  "pollS3",
			RouteNames: []string{"set_rule"},
			Routine:    PollS3,
		},
		&BlockTemplate{
			BlockType:  "tofile",
			RouteNames: []string{"set_rule"},
			Routine:    ToFile,
		},
		&BlockTemplate{
			BlockType:  "filter",
			RouteNames: []string{"set_rule", "get_rule"},
			Routine:    Filter,
		},
		&BlockTemplate{
			BlockType:  "postto",
			RouteNames: []string{"in"},
			Routine:    PostTo,
		},
		&BlockTemplate{
			BlockType:  "timeseries",
			RouteNames: []string{"set_rule", "timeseries"},
			Routine:    Timeseries,
		},
	}

	for _, t := range templates {
		Library[t.BlockType] = t
	}
}
