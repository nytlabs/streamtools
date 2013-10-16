package streamtools

import (
    //"github.com/bitly/go-simplejson"
    "log"
)

const (
    InChan = iota
    OutChan = iota
)

// The hub collects to gether all the blocks and connections
type hub struct {
	connectionMap map[string]*Connection
	blockMap      map[string]*Block
}

func (self *hub) CreateBlock(blockType string){
    block := NewBlock(blockType)
    log.Println(block)
    //block.
    
    //self.blockMap[blockType + "_" + block.ID] = block
    log.Println(block.getID())
    go block.blockRoutine()
}

/*func (self *hub) CreateConnection(chanName string){
    if connection, ok := self.connectionMap[chanName]; ok {
        log.Println("chan already exists")
    } else {
        self.connectionMap[chanName] := NewConnection(chanName)
        log.Println("created " + chanName)
    }
}

func (self *hub) SetBlockChan(chanName string, blockID string, ChanType int){
    if block, ok := self.blockMap[blockID]; ok{
        if blockChan, ok := self.connectionMap[chanName]; ok {
            block.SetChan(blockChan, ChanType)
        } else {
            self.CreateConnection(chanName)
            block.SetChan(connectionMap[chanName], ChanType)
        }
    } else {
        log.Println("blockID not found")
    }
}*/

/*func (self *hub) DestroyBlock(blockID string){
    if block, ok := self.blockMap[blockID]; ok{
        delete(self.blockMap, blockID)
    }
}*/