package daemon

import (
    "errors"
    "fmt"
)

type BlockInfo struct {
    Id string
    Type string
    X float64
    Y float64
}

type ConnectionInfo struct {
    Id string
    FromId string
    ToId string
    ToRoute string
}

type BlockManager struct {
    blockMap map[string]*BlockInfo
    connMap map[string]*ConnectionInfo
}

func NewBlockManager() *BlockManager {
    return &BlockManager{
        blockMap: make(map[string]*BlockInfo),
        connMap: make(map[string]*ConnectionInfo),
    }
}

func (b *BlockManager) Create(id string, blockType string, x float64, y float64) error {
    _, ok := b.blockMap[id]
    if ok {
        return errors.New(fmt.Sprintf("block ID: %s already in use.", id))
    }
    return nil
}

func (b *BlockManager) Send(id string, route string, msg interface{}) error {
    _, ok := b.blockMap[id]
    if !ok {
        return errors.New(fmt.Sprintf("block ID: %s does not exist", id))
    }

    return nil
}

func (b *BlockManager) Query(id string, route string) error {
    _, ok := b.blockMap[id]
    if !ok {
        return errors.New(fmt.Sprintf("block ID: %s does not exist", id))
    }

    return nil
}

func (b *BlockManager) Connect(id string, fromId string, toId string, toRoute string) error {
    _, ok := b.connMap[id]
    if ok {
        return errors.New(fmt.Sprintf("connection ID: %s already in use.", id))
    }
    _, ok = b.blockMap[fromId]
    if !ok {
        return errors.New(fmt.Sprintf("block ID: %s does not exist", id))
    }
    _, ok = b.blockMap[toId]
    if !ok {
        return errors.New(fmt.Sprintf("block ID: %s does not exist", id))
    }


    return nil
}

func (b *BlockManager) GetBlock(id string) *BlockInfo, error {
    _, ok := b.blockMap[id]
    if !ok {
        return errors.New(fmt.Sprintf("block ID: %s does not exist", id))
    }
    return nil, nil
}

func (b *BlockManager) GetConnection(id string) *ConnectionInfo, error {
    _, ok := b.connMap[id]
    if !ok {
        return errors.New(fmt.Sprintf("connection ID: %s does not exist", id))
    }
    return nil, nil
}

func (b *BlockManager) Delete(id string) error {
    _, ok := b.blockMap[id]
    if !ok {
        return errors.New(fmt.Sprintf("block ID: %s does not exist", id))
    }

    return nil
}

func (b *BlockManager) Disconnect(id string) error {
    _, ok := b.connMap[id]
    if !ok {
        return errors.New(fmt.Sprintf("connection ID: %s does not exist", id))
    }

    return nil
}


func (b *BlockManager) ListBlocks() []*BlockInfo {

    return nil
}

func (b *BlockManager) ListConnections() []*ConnectionInfo {

    return nil
}