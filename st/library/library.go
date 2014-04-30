// +build !arm

package library

import (
	"github.com/nytlabs/streamtools/st/blocks"
)

var Blocks = map[string]func() blocks.BlockInterface{
	"count":              NewCount,
	"toggle":             NewToggle,
	"movingaverage":      NewMovingAverage,
	"ticker":             NewTicker,
	"fromnsq":            NewFromNSQ,
	"fromamqp":           NewFromAMQP,
	"fromhttpstream":     NewFromHTTPStream,
	"fromHTTPGetRequest": NewFromHTTPGetRequest,
	"fromsqs":            NewFromSQS,
	"frompost":           NewFromPost,
	"fromfile":           NewFromFile,
	"fromemail":          NewFromEmail,
	"tonsq":              NewToNSQ,
	"toamqp":             NewToAMQP,
	"toelasticsearch":    NewToElasticsearch,
	"toemail":            NewToEmail,
	"tofile":             NewToFile,
	"tolog":              NewToLog,
	"tobeanstalkd":       NewToBeanstalkd,
	"tomongodb":          NewToMongoDB,
	"toHTTPGetRequest":   NewToHTTPGetRequest,
	"mask":               NewMask,
	"filter":             NewFilter,
	"sync":               NewSync,
	"queue":              NewQueue,
	"unpack":             NewUnpack,
	"packbyinterval":     NewPackByInterval,
	"packbyvalue":        NewPackByValue,
	"packbycount":        NewPackByCount,
	"parsexml":           NewParseXML,
	"set":                NewSet,
	"cache":              NewCache,
	"join":               NewJoin,
	"kullbackleibler":    NewKullbackLeibler,
	"learn":              NewLearn,
	"logisticModel":      NewLogisticModel,
	"linearModel":        NewLinearModel,
	"gethttp":            NewGetHTTP,
	"gaussian":           NewGaussian,
	"zipf":               NewZipf,
	"poisson":            NewPoisson,
	"categorical":        NewCategorical,
	"map":                NewMap,
	"histogram":          NewHistogram,
	"timeseries":         NewTimeseries,
	"fromwebsocket":      NewFromWebsocket,
	"tonsqmulti":         NewToNSQMulti,
	"fromudp":            NewFromUDP,
	"dedupe":             NewDeDupe,
	"javascript":         NewJavascript,
	"fft":                NewFFT,
}

var BlockDefs = map[string]*blocks.BlockDef{}

func Start() {
	for k, newBlock := range Blocks {
		b := newBlock()
		b.Build(blocks.BlockChans{nil, nil, nil, nil, nil, nil, nil})
		b.Setup()
		BlockDefs[k] = b.GetDef()
	}
}
