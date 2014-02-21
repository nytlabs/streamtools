package library

import(
    "github.com/nytlabs/streamtools/st/blocks"
)

var Blocks = map[string] func() blocks.BlockInterface {
    "count": NewCount,
    "ticker": NewTicker,
    "fromnsq": NewFromNSQ,
    "tonsq": NewToNSQ,
    "tofile": NewToFile,
    "tolog": NewToLog,
    "mask": NewMask,
    "filter": NewFilter,
    "fromsqs": NewFromSQS,
    "sync": NewSync,
    "gethttp": NewGetHTTP,
    "fromhttpstream": NewFromHTTPStream,
}

var BlockDefs = map[string]*blocks.BlockDef{}

func Start(){
    for k, newBlock := range Blocks {
        b := newBlock()
        b.Build(blocks.BlockChans{nil,nil,nil,nil,nil,nil})
        b.Setup()
        BlockDefs[k] = b.GetDef()
    }
}