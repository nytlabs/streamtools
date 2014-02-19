package library

import(
    "github.com/nytlabs/streamtools/st/blocks"
)

var Blocks = map[string] func() blocks.BlockInterface {
    "count": NewCount,
}

var BlockDefs = map[string]*blocks.BlockDef{}

func BuildLibrary(){
    for k, newBlock := range Blocks {
        b := newBlock()
        b.Build(blocks.BlockChans{nil,nil,nil,nil,nil,nil})
        b.Setup()
        BlockDefs[k] = b.GetDef()
    }
}