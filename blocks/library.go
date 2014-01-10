package blocks

import (
	"encoding/json"
)

var (
	Library     BlockLibrary
	LibraryBlob string
)

// A block library is a collection of possible block templates
type BlockLibrary map[string]*BlockTemplate

func BuildLibrary() {
	Library = make(map[string]*BlockTemplate)

	templates := []*BlockTemplate{
		&BlockTemplate{
			BlockType:  "connection",
			RouteNames: []string{"last_message", "rate"},
			Routine:    Connection,
		},
		////////// TESTING BLOCKS
		&BlockTemplate{
			BlockType:  "blocked",
			RouteNames: []string{"get_rule"},
			Routine:    Blocked,
		},
		////////////////////
		&BlockTemplate{
			BlockType:  "ticker",
			RouteNames: []string{"set_rule", "get_rule"},
			Routine:    Ticker,
		},
		&BlockTemplate{
			BlockType:  "tolog",
			RouteNames: []string{},
			Routine:    ToLog,
		},
		&BlockTemplate{
			BlockType:  "random",
			RouteNames: []string{"set_rule", "get_rule"},
			Routine:    Random,
		},
		&BlockTemplate{
			BlockType:  "count",
			RouteNames: []string{"set_rule", "get_rule", "count", "poll"},
			Routine:    Count,
		},
		&BlockTemplate{
			BlockType:  "mask",
			RouteNames: []string{"set_rule", "get_rule"},
			Routine:    Mask,
		},
		&BlockTemplate{
			BlockType:  "unpack",
			RouteNames: []string{"set_rule", "get_rule"},
			Routine:    Unpack,
		},
		&BlockTemplate{
			BlockType:  "sync",
			RouteNames: []string{"set_rule", "get_rule"},
			Routine:    Sync,
		},
		&BlockTemplate{
			BlockType:  "postValue",
			RouteNames: []string{"set_rule", "get_rule"},
			Routine:    PostValue,
		},
		&BlockTemplate{
			BlockType:  "toPost",
			RouteNames: []string{"set_rule", "get_rule"},
			Routine:    Post,
		},
		&BlockTemplate{
			BlockType:  "getHTTP",
			RouteNames: []string{"set_rule", "get_rule"},
			Routine:    GetHTTP,
		},
		&BlockTemplate{
			BlockType:  "date",
			RouteNames: []string{"set_rule", "get_rule"},
			Routine:    Date,
		},
		&BlockTemplate{
			BlockType:  "NSQStream",
			RouteNames: []string{"set_rule", "get_rule"},
			Routine:    NSQStream,
		},
		&BlockTemplate{
			BlockType:  "getS3",
			RouteNames: []string{"set_rule", "get_rule"},
			Routine:    GetS3,
		},
		&BlockTemplate{
			BlockType:  "listS3",
			RouteNames: []string{"set_rule", "get_rule"},
			Routine:    ListS3,
		},
		&BlockTemplate{
			BlockType:  "tofile",
			RouteNames: []string{"set_rule", "get_rule"},
			Routine:    ToFile,
		},
		&BlockTemplate{
			BlockType:  "toNSQ",
			RouteNames: []string{"set_rule", "get_rule"},
			Routine:    ToNSQ,
		},
		&BlockTemplate{
			BlockType:  "filter",
			RouteNames: []string{"set_rule", "get_rule"},
			Routine:    Filter,
		},
		&BlockTemplate{
			BlockType:  "postHTTP",
			RouteNames: []string{"in"},
			Routine:    PostHTTP,
		},
		&BlockTemplate{
			BlockType:  "timeseries",
			RouteNames: []string{"set_rule", "timeseries", "get_rule"},
			Routine:    Timeseries,
		},
		&BlockTemplate{
			BlockType:  "histogram",
			RouteNames: []string{"set_rule", "histogram", "get_rule", "poll"},
			Routine:    Histogram,
		},
		&BlockTemplate{
			BlockType:  "groupHistogram",
			RouteNames: []string{"set_rule", "histogram", "get_rule", "list"},
			Routine:    GroupHistogram,
		},
		&BlockTemplate{
			BlockType:  "pack",
			RouteNames: []string{"set_rule", "get_rule"},
			Routine:    Pack,
		},
		&BlockTemplate{
			BlockType:  "mean",
			RouteNames: []string{"set_rule", "get_rule", "avg"},
			Routine:    Mean,
		},
		&BlockTemplate{
			BlockType:  "sd",
			RouteNames: []string{"set_rule", "get_rule", "sd"},
			Routine:    Sd,
		},
		&BlockTemplate{
			BlockType:  "var",
			RouteNames: []string{"set_rule", "get_ruel", "var"},
			Routine:    Var,
		},
		&BlockTemplate{
			BlockType:  "HTTPStream",
			RouteNames: []string{"set_rule", "get_rule"},
			Routine:    HTTPStream,
		},
		&BlockTemplate{
			BlockType:  "SQSStream",
			RouteNames: []string{"set_rule", "get_rule"},
			Routine:    SQSStream,
		},
		&BlockTemplate{
			BlockType:  "map",
			RouteNames: []string{"set_rule", "get_rule"},
			Routine:    Map,
		},
		&BlockTemplate{
			BlockType:  "linearModel",
			RouteNames: []string{"set_rule", "get_rule"},
			Routine:    LinearModel,
		},
	}

	libraryList := []map[string]interface{}{}
	for _, t := range templates {
		blockItem := make(map[string]interface{})
		blockItem["BlockType"] = t.BlockType
		blockItem["RouteNames"] = t.RouteNames
		libraryList = append(libraryList, blockItem)

		Library[t.BlockType] = t
	}

	blob, _ := json.Marshal(libraryList)
	LibraryBlob = string(blob)
}
