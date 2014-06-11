package library

import (
	"github.com/nytlabs/streamtools/st/blocks"
)

var Blocks = map[string]func() blocks.BlockInterface{
	"analogPin":          NewAnalogPin,
	"digitalpin":         NewDigitalPin,
	"todigitalpin":       NewToDigitalPin,
	"bang":               NewBang,
	"cache":              NewCache,
	"categorical":        NewCategorical,
	"count":              NewCount,
	"dedupe":             NewDeDupe,
	"fft":                NewFFT,
	"filter":             NewFilter,
	"fromamqp":           NewFromAMQP,
	"fromemail":          NewFromEmail,
	"fromfile":           NewFromFile,
	"fromHTTPGetRequest": NewFromHTTPGetRequest,
	"fromhttpstream":     NewFromHTTPStream,
	"fromnsq":            NewFromNSQ,
	"frompost":           NewFromPost,
	"fromsqs":            NewFromSQS,
	"fromwebsocket":      NewFromWebsocket,
	"fromudp":            NewFromUDP,
	"gaussian":           NewGaussian,
	"gethttp":            NewGetHTTP,
	"histogram":          NewHistogram,
	"join":               NewJoin,
	"kullbackleibler":    NewKullbackLeibler,
	"learn":              NewLearn,
	"linearModel":        NewLinearModel,
	"logisticModel":      NewLogisticModel,
	"map":                NewMap,
	"mask":               NewMask,
	"movingaverage":      NewMovingAverage,
	"packbycount":        NewPackByCount,
	"packbyinterval":     NewPackByInterval,
	"packbyvalue":        NewPackByValue,
	"parsecsv":           NewParseCSV,
	"parsexml":           NewParseXML,
	"poisson":            NewPoisson,
	"javascript":         NewJavascript,
	"queue":              NewQueue,
	"redis":              NewRedis,
	"set":                NewSet,
	"sync":               NewSync,
	"ticker":             NewTicker,
	"timeseries":         NewTimeseries,
	"toamqp":             NewToAMQP,
	"tobeanstalkd":       NewToBeanstalkd,
	"toelasticsearch":    NewToElasticsearch,
	"toemail":            NewToEmail,
	"tofile":             NewToFile,
	"toggle":             NewToggle,
	"toHTTPGetRequest":   NewToHTTPGetRequest,
	"tolog":              NewToLog,
	"tomongodb":          NewToMongoDB,
	"tonsq":              NewToNSQ,
	"tonsqmulti":         NewToNSQMulti,
	"unpack":             NewUnpack,
	"webRequest":         NewWebRequest,
	"zipf":               NewZipf,
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
