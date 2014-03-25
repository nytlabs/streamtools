package library

import (
	"github.com/nytlabs/streamtools/st/blocks"
)

var Blocks = map[string]func() blocks.BlockInterface{
	"count":           NewCount,
	"movingaverage":   NewMovingAverage,
	"analogpin":       NewAnalogPin,
	"ticker":          NewTicker,
	"fromnsq":         NewFromNSQ,
	"fromhttpstream":  NewFromHTTPStream,
	"fromsqs":         NewFromSQS,
	"frompost":        NewFromPost,
	"tonsq":           NewToNSQ,
	"toelasticsearch": NewToElasticsearch,
	"towebsocket":     NewToWebsocket,
	"tofile":          NewToFile,
	"tolog":           NewToLog,
	"tobeanstalkd":    NewToBeanstalkd,
	"mask":            NewMask,
	"filter":          NewFilter,
	"sync":            NewSync,
	"unpack":          NewUnpack,
	"pack":            NewPack,
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
