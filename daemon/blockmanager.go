package daemon

import (
	"errors"
	"fmt"
	"github.com/nytlabs/streamtools/blocks"
	"strconv"
)

// so i don't forget:
// all blocks should clear a message's route
// connects should set messages route.

type BlockInfo struct {
	Id   string
	Type string
	Rule interface{}
	X    float64
	Y    float64
	in   chan interface{}
}

type ConnectionInfo struct {
	Id      string
	FromId  string
	ToId    string
	ToRoute string
	in      chan interface{}
}

type BlockManager struct {
	blockMap map[string]*BlockInfo
	connMap  map[string]*ConnectionInfo
	genId    chan string
}

func IDService(idChan chan string) {
	i := 1
	for {
		id := strconv.Itoa(i)
		idChan <- id
		i += 1
	}
}

func NewBlockManager() *BlockManager {
	idChan := make(chan string)
	go IDService(idChan)
	blocks.BuildLibrary()
	return &BlockManager{
		blockMap: make(map[string]*BlockInfo),
		connMap:  make(map[string]*ConnectionInfo),
		genId:    idChan,
	}
}

func (b *BlockManager) getID() string {
	id := <-b.genId
	ok := b.idExists(id)
	for ok {
		id = <-b.genId
		ok = b.idExists(id)
	}
	return id
}

func (b *BlockManager) idExists(id string) bool {
	_, okB := b.blockMap[id]
	_, okC := b.connMap[id]
	return okB || okC
}

func (b *BlockManager) Create(block *BlockInfo) (*BlockInfo, error) {
	if block == nil {
		return nil, errors.New("Cannot create: no block data.")
	}

	// create ID if there is none
	if block.Id == "" {
		block.Id = b.getID()
	}

	// make sure ID doesn't already exist
	if b.idExists(block.Id) {
		return nil, errors.New("Cannot create: id already exists")
	}

	// check block Type
	_, ok := blocks.Library[block.Type]
	if !ok {
		return nil, errors.New("Cannot create: invalid block type")
	}

	// go blockroutine create block here
	b.blockMap[block.Id] = block

	// if rule != nil
	// do a send on the rule.

	return block, nil
}

func (b *BlockManager) Send(id string, route string, msg interface{}) error {
	_, ok := b.blockMap[id]
	if !ok {
		return errors.New(fmt.Sprintf("block ID: %s does not exist", id))
	}
	// send message to block here

	return nil
}

func (b *BlockManager) Query(id string, route string) (interface{}, error) {
	_, ok := b.blockMap[id]
	if !ok {
		return nil, errors.New(fmt.Sprintf("block ID: %s does not exist", id))
	}
	// send qury to block here

	return nil, nil
}

func (b *BlockManager) Connect(conn *ConnectionInfo) (*ConnectionInfo, error) {
	if conn == nil {
		return nil, errors.New("Cannot create: no connection data.")
	}

	// create ID if there is none
	if conn.Id == "" {
		conn.Id = b.getID()
	}

	// make sure ID doesn't already exist
	if b.idExists(conn.Id) {
		return nil, errors.New("Cannot create: id already exists")
	}

	// check to see if the blocks that we are attaching to exist
	fromExists := b.idExists(conn.FromId)
	if !fromExists {
		return nil, errors.New("Cannot create: FromId ID does not exist")
	}

	toExists := b.idExists(conn.ToId)
	if !toExists {
		return nil, errors.New("Cannot create: ToId ID does not exist")
	}

	// go blockroutine create block here
	b.connMap[conn.Id] = conn

	return conn, nil
}

func (b *BlockManager) GetBlock(id string) (*BlockInfo, error) {
	block, ok := b.blockMap[id]
	if !ok {
		return nil, errors.New(fmt.Sprintf("block ID: %s does not exist", id))
	}

	// retrieve block's and set it here.
	block.Rule = "test" // "retrieve fresh rule from the block here..."

	return block, nil
}

func (b *BlockManager) GetConnection(id string) (*ConnectionInfo, error) {
	_, ok := b.connMap[id]
	if !ok {
		return nil, errors.New(fmt.Sprintf("connection ID: %s does not exist", id))
	}
	return b.connMap[id], nil
}

func (b *BlockManager) DeleteBlock(id string) error {
	_, ok := b.blockMap[id]
	if !ok {
		return errors.New(fmt.Sprintf("block ID: %s does not exist", id))
	}

	// turn off block here
	// close channels, whatever.

	delete(b.blockMap, id)

	return nil
}

func (b *BlockManager) DeleteConnection(id string) error {
	_, ok := b.connMap[id]
	if !ok {
		return errors.New(fmt.Sprintf("connection ID: %s does not exist", id))
	}

	// call disconnecting stuff here
	// remove channel from FromBlock, etc
	// turn off connection block
	delete(b.connMap, id)

	return nil
}

func (b *BlockManager) ListBlocks() []*BlockInfo {
	i := 0
	blocks := make([]*BlockInfo, len(b.blockMap), len(b.blockMap))
	for k, _ := range b.blockMap {
		v, err := b.GetBlock(k)
		if err != nil {
			continue
		}
		blocks[i] = v
		i++
	}

	return blocks
}

func (b *BlockManager) ListConnections() []*ConnectionInfo {
	i := 0
	conns := make([]*ConnectionInfo, len(b.connMap), len(b.connMap))
	for _, v := range b.connMap {
		conns[i] = v
		i++
	}
	return conns
}
