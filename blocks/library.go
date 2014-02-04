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

		// generators : genX
		&BlockTemplate{
			BlockType:  "genTicker",
			RouteNames: []string{"set_rule", "get_rule"},
			Routine:    GenTicker,
		},
		&BlockTemplate{
			BlockType:  "genRandom",
			RouteNames: []string{"set_rule", "get_rule"},
			Routine:    GenRandom,
		},

		/// sources : fromX
		&BlockTemplate{
			BlockType:  "fromNSQ",
			RouteNames: []string{"set_rule", "get_rule"},
			Routine:    FromNSQ,
		},
		&BlockTemplate{
			BlockType:  "fromHTTPStream",
			RouteNames: []string{"set_rule", "get_rule"},
			Routine:    FromHTTPStream,
		},
		&BlockTemplate{
			BlockType:  "fromSQS",
			RouteNames: []string{"set_rule", "get_rule"},
			Routine:    FromSQS,
		},
		&BlockTemplate{
			BlockType:  "fromPost",
			RouteNames: []string{"in"},
			Routine:    FromPost,
		},

		/// sinks: toX
		&BlockTemplate{
			BlockType:  "toLog",
			RouteNames: []string{},
			Routine:    ToLog,
		},
		&BlockTemplate{
			BlockType:  "toWebsocket",
			RouteNames: []string{"set_rule", "get_rule"},
			Routine:    ToWebsocket,
		},
		&BlockTemplate{
			BlockType:  "toElasticsearch",
			RouteNames: []string{"set_rule", "get_rule"},
			Routine:    ToElasticsearch,
		},
		&BlockTemplate{
			BlockType:  "toFile",
			RouteNames: []string{"set_rule", "get_rule"},
			Routine:    ToFile,
		},
		&BlockTemplate{
			BlockType:  "toNSQ",
			RouteNames: []string{"set_rule", "get_rule"},
			Routine:    ToNSQ,
		},

		/// state blocks
		&BlockTemplate{
			BlockType:  "count",
			RouteNames: []string{"set_rule", "get_rule", "count", "poll"},
			Routine:    Count,
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
			BlockType:  "learn",
			RouteNames: []string{"set_rule", "get_rule", "state", "poll"},
			Routine:    Learn,
		},

		/// transfer blocks
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
			BlockType:  "postHTTP",
			RouteNames: []string{"set_rule", "get_rule"},
			Routine:    PostHTTP,
		},
		&BlockTemplate{
			BlockType:  "getHTTP",
			RouteNames: []string{"set_rule", "get_rule"},
			Routine:    GetHTTP,
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
			BlockType:  "filter",
			RouteNames: []string{"set_rule", "get_rule"},
			Routine:    Filter,
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
			BlockType:  "movingAverage",
			RouteNames: []string{"set_rule", "get_rule", "moving_average", "poll"},
			Routine:    MovingAverage,
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
