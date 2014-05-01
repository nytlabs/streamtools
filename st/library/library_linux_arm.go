package library

import (
	"github.com/nytlabs/streamtools/st/blocks"
)

var Blocks = map[string]func() blocks.BlockInterface{
	"count":           NewCount,
	"toggle":          NewToggle,
	"movingaverage":   NewMovingAverage,
	"fft":             NewFFT,
	"ticker":          NewTicker,
	"analogPin":       NewAnalogPin,
	"digitalpin":      NewDigitalPin,
	"todigitalpin":    NewToDigitalPin,
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
	"packbyinterval":  NewPackByInterval,
	"packbyvalue":     NewPackByValue,
	"packbycount":     NewPackByCount,
	"parsexml":        NewParseXML,
	"set":             NewSet,
	"cache":           NewCache,
	"join":            NewJoin,
	"gethttp":         NewGetHTTP,
	"gaussian":        NewGaussian,
	"zipf":            NewZipf,
	"poisson":         NewPoisson,
	"categorical":     NewCategorical,
	"map":             NewMap,
	"histogram":       NewHistogram,
	"timeseries":      NewTimeseries,
	"fromwebsocket":   NewFromWebsocket,
	"tonsqmulti":      NewToNSQMulti,
	"fromudp":         NewFromUDP,
	"dedupe":          NewDeDupe,
	"javascript":      NewJavascript,
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
