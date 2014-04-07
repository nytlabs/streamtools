package library

import (
	"github.com/nytlabs/streamtools/st/blocks"
)

var Blocks = map[string]func() blocks.BlockInterface{
	"count":           NewCount,
	"toggle":          NewToggle,
	"movingaverage":   NewMovingAverage,
	"analogpin":       NewAnalogPin,
	"ticker":          NewTicker,
	"fromnsq":         NewFromNSQ,
	"fromhttpstream":  NewFromHTTPStream,
	"fromsqs":         NewFromSQS,
	"frompost":        NewFromPost,
	"fromfile":        NewFromFile,
	"tonsq":           NewToNSQ,
	"toelasticsearch": NewToElasticsearch,
	"tofile":          NewToFile,
	"tolog":           NewToLog,
	"tobeanstalkd":    NewToBeanstalkd,
	"tomongodb":       NewToMongoDB,
	"mask":            NewMask,
	"filter":          NewFilter,
	"sync":            NewSync,
	"queue":           NewQueue,
	"unpack":          NewUnpack,
	"pack":            NewPack,
	"parsexml":        NewParseXML,
	"set":             NewSet,
	"join":            NewJoin,
	"gethttp":         NewGetHTTP,
	"gaussian":        NewGaussian,
	"zipf":            NewZipf,
	"poisson":         NewPoisson,
	"map":             NewMap,
	"histogram":       NewHistogram,
	"timeseries":      NewTimeseries,
	"fromwebsocket":   NewFromWebsocket,
	"tonsqmulti":      NewToNSQMulti,
	"fromudp":         NewFromUDP,
}

var BlockDefs = map[string]*blocks.BlockDef{}

func Start() {
	for k, newBlock := range Blocks {
		b := newBlock()
		b.Build(blocks.BlockChans{nil, nil, nil, nil, nil, nil})
		b.Setup()
		BlockDefs[k] = b.GetDef()
	}
}
