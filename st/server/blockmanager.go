package server

import (
	"errors"
	"fmt"
	"github.com/nytlabs/streamtools/st/blocks"
	"github.com/nytlabs/streamtools/st/library"
	"net/url"
	"strconv"
	"sync"
	"time"
)

type BlockInfo struct {
	Id       string
	Type     string
	Rule     interface{}
	Position *Coords
	chans    blocks.BlockChans
}

type ConnectionInfo struct {
	Id      string
	FromId  string
	ToId    string
	ToRoute string
	chans   blocks.BlockChans
}

type Coords struct {
	X float64
	Y float64
}

type BlockManager struct {
	blockMap map[string]*BlockInfo
	connMap  map[string]*ConnectionInfo
	genId    chan string
	Mu       *sync.Mutex
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
	return &BlockManager{
		blockMap: make(map[string]*BlockInfo),
		connMap:  make(map[string]*ConnectionInfo),
		genId:    idChan,
		Mu:       &sync.Mutex{},
	}
}

func (b *BlockManager) GetId() string {
	id := <-b.genId
	ok := b.IdExists(id)
	for ok {
		id = <-b.genId
		ok = b.IdExists(id)
	}
	return id
}

func (b *BlockManager) IdExists(id string) bool {
	_, okB := b.blockMap[id]
	_, okC := b.connMap[id]
	return okB || okC
}

func (b *BlockManager) IdSafe(id string) bool {
	return url.QueryEscape(id) == id && id != "DAEMON"
}

func (b *BlockManager) Create(blockInfo *BlockInfo) (*BlockInfo, error) {
	if blockInfo == nil {
		return nil, errors.New(fmt.Sprintf("Cannot create block %s: no block data.", blockInfo.Id))
	}

	// check to see if the ID is OK
	if !b.IdSafe(blockInfo.Id) {
		return nil, errors.New(fmt.Sprintf("Cannot create block %s: invalid id", blockInfo.Id))
	}

	// create ID if there is none
	if blockInfo.Id == "" {
		blockInfo.Id = b.GetId()
	}

	// make sure ID doesn't already exist
	if b.IdExists(blockInfo.Id) {
		return nil, errors.New(fmt.Sprintf("Cannot create block %s: id already exists", blockInfo.Id))
	}

	// give the block a position if it doesn't have one.
	if blockInfo.Position == nil {
		blockInfo.Position = &Coords{
			X: 0,
			Y: 0,
		}
	}

	_, ok := library.Blocks[blockInfo.Type]
	if !ok {
		return nil, errors.New(fmt.Sprintf("Cannot create block %s: invalid block type %s", blockInfo.Id, blockInfo.Type))
	}

	// create the block
	newBlock := library.Blocks[blockInfo.Type]()

	newBlockChans := blocks.BlockChans{
		InChan:         make(chan *blocks.Msg),
		QueryChan:      make(chan *blocks.QueryMsg),
		QueryParamChan: make(chan *blocks.QueryParamMsg),
		AddChan:        make(chan *blocks.AddChanMsg),
		DelChan:        make(chan *blocks.Msg),
		ErrChan:        make(chan error),
		IdChan:         make(chan string),
		QuitChan:       make(chan bool),
	}

	newBlock.SetId(blockInfo.Id)
	newBlock.Build(newBlockChans)
	go blocks.BlockRoutine(newBlock)

	// save state
	blockInfo.chans = newBlockChans
	b.blockMap[blockInfo.Id] = blockInfo

	if blockInfo.Rule != nil {
		err := b.Send(blockInfo.Id, "rule", blockInfo.Rule)
		if err != nil {
			return nil, err
		}
	} else {
		b.updateRule(blockInfo.Id)
	}

	return blockInfo, nil
}

func (b *BlockManager) UpdateBlockPosition(id string, coord *Coords) (*BlockInfo, error) {
	block, ok := b.blockMap[id]
	if !ok {
		return nil, errors.New(fmt.Sprintf("Cannot update block %s: does not exist", id))
	}

	block.Position = coord

	return block, nil
}

func (b *BlockManager) Send(id string, route string, msg interface{}) error {
	_, ok := b.blockMap[id]
	if !ok {
		return errors.New(fmt.Sprintf("Cannot send to block %s: does not exist", id))
	}
	// send message to block here
	b.blockMap[id].chans.InChan <- &blocks.Msg{
		Msg:   msg,
		Route: route,
	}

	return nil
}

func (b *BlockManager) QueryBlock(id string, route string) (interface{}, error) {
	_, ok := b.blockMap[id]
	if !ok {
		return nil, errors.New(fmt.Sprintf("Cannot query block %s: does not exist", id))
	}
	var returnToSender blocks.MsgChan
	returnToSender = make(chan interface{})
	b.blockMap[id].chans.QueryChan <- &blocks.QueryMsg{
		Route:   route,
		MsgChan: returnToSender,
	}
	timeout := time.NewTimer(1 * time.Second)
	select {
	case q := <-returnToSender:
		return q, nil
	case <-timeout.C:
		return nil, errors.New(fmt.Sprintf("Cannot query block %s: timeout", id))
	}
}

func (b *BlockManager) QueryParamBlock(id string, route string, params url.Values) (interface{}, error) {
	_, ok := b.blockMap[id]
	if !ok {
		return nil, errors.New(fmt.Sprintf("Cannot query block %s: does not exist", id))
	}
	var returnToSender blocks.MsgChan
	returnToSender = make(chan interface{})
	b.blockMap[id].chans.QueryParamChan <- &blocks.QueryParamMsg{
		Route:    route,
		RespChan: returnToSender,
		Params:   params,
	}
	timeout := time.NewTimer(1 * time.Second)
	select {
	case q := <-returnToSender:
		return q, nil
	case <-timeout.C:
		return nil, errors.New(fmt.Sprintf("Cannot query block %s: timeout", id))
	}
}

func (b *BlockManager) QueryConnection(id string, route string) (interface{}, error) {
	_, ok := b.connMap[id]
	if !ok {
		return nil, errors.New(fmt.Sprintf("Cannot query block %s: does not exist", id))
	}

	var returnToSender blocks.MsgChan
	returnToSender = make(chan interface{})
	msg := &blocks.QueryMsg{
		Route:   route,
		MsgChan: returnToSender,
	}
	b.connMap[id].chans.QueryChan <- msg
	q := <-returnToSender

	return q, nil
}

func (b *BlockManager) Connect(connInfo *ConnectionInfo) (*ConnectionInfo, error) {
	if connInfo == nil {
		return nil, errors.New("Cannot create: no connection data.")
	}

	// check to see if the ID is OK
	if !b.IdSafe(connInfo.Id) {
		return nil, errors.New(fmt.Sprintf("Cannot create block %s: invalid id", connInfo.Id))
	}

	// create ID if there is none
	if connInfo.Id == "" {
		connInfo.Id = b.GetId()
	}

	// make sure ID doesn't already exist
	if b.IdExists(connInfo.Id) {
		return nil, errors.New(fmt.Sprintf("Cannot create connection %s: id already exists", connInfo.Id))
	}

	// check to see if the blocks that we are attaching to exist
	fromExists := b.IdExists(connInfo.FromId)
	if !fromExists {
		return nil, errors.New(fmt.Sprintf("Cannot create connection %s: FromId block does not exist", connInfo.Id))
	}

	toExists := b.IdExists(connInfo.ToId)
	if !toExists {
		return nil, errors.New(fmt.Sprintf("Cannot create connection %s: ToId ID does not exist", connInfo.Id))
	}

	// create connection info for server
	// and create connection routine
	newConn := &blocks.Connection{
		ToRoute: connInfo.ToRoute,
	}

	newConnChans := blocks.BlockChans{
		InChan:         make(chan *blocks.Msg),
		QueryChan:      make(chan *blocks.QueryMsg),
		QueryParamChan: make(chan *blocks.QueryParamMsg),
		AddChan:        make(chan *blocks.AddChanMsg),
		DelChan:        make(chan *blocks.Msg),
		ErrChan:        make(chan error),
		QuitChan:       make(chan bool),
	}

	newConn.SetId(connInfo.Id)
	newConn.Build(newConnChans)
	go blocks.ConnectionRoutine(newConn)

	connInfo.chans = newConnChans
	b.connMap[connInfo.Id] = connInfo

	// ask to connect the blocks together
	b.blockMap[connInfo.FromId].chans.AddChan <- &blocks.AddChanMsg{
		Route:   connInfo.Id,
		Channel: connInfo.chans.InChan,
	}

	b.connMap[connInfo.Id].chans.AddChan <- &blocks.AddChanMsg{
		Route:   connInfo.ToId,
		Channel: b.blockMap[connInfo.ToId].chans.InChan,
	}

	return connInfo, nil
}

func (b *BlockManager) GetSocket(fromId string) (chan *blocks.Msg, string, error) {
	_, ok := b.blockMap[fromId]
	if !ok {
		return nil, "", errors.New(fmt.Sprintf("Cannot recieve from block %s: does not exist", fromId))
	}

	wsChan := make(chan *blocks.Msg)
	id := b.GetId()

	b.blockMap[fromId].chans.AddChan <- &blocks.AddChanMsg{
		Route:   id,
		Channel: wsChan,
	}

	return wsChan, id, nil
}

func (b *BlockManager) DeleteSocket(blockId string, connId string) error {
	if _, ok := b.blockMap[blockId]; ok {
		b.blockMap[blockId].chans.DelChan <- &blocks.Msg{
			Route: connId,
		}
	}
	return nil
}

func (b *BlockManager) updateRule(id string) {
	rule := false
	block := b.blockMap[id]
	for _, b := range library.BlockDefs[block.Type].QueryRoutes {
		rule = b == "rule"
		if rule {
			break
		}
	}

	if rule {
		q, err := b.QueryBlock(id, "rule")
		if err != nil {
			return
		}

		block.Rule = q
	}
}

func (b *BlockManager) GetBlock(id string) (*BlockInfo, error) {
	block, ok := b.blockMap[id]
	if !ok {
		return nil, errors.New(fmt.Sprintf("Cannot get block %s: does not exist", id))
	}

	b.updateRule(id)

	return block, nil
}

func (b *BlockManager) GetConnection(id string) (*ConnectionInfo, error) {
	_, ok := b.connMap[id]
	if !ok {
		return nil, errors.New(fmt.Sprintf("Cannot get connection %s: does not exist", id))
	}
	return b.connMap[id], nil
}

func (b *BlockManager) DeleteBlock(id string) ([]string, error) {
	var delIds []string

	_, ok := b.blockMap[id]
	if !ok {
		return nil, errors.New(fmt.Sprintf("Cannot delete block %s: does not exist", id))
	}

	// delete connections that reference this block
	for _, c := range b.connMap {
		if c.FromId == id {
			delFromId, err := b.DeleteConnection(c.Id)
			if err != nil {
				return nil, errors.New(fmt.Sprintf("Cannot delete block %s: FromId %s does not exist", id, c.FromId))
			}
			delIds = append(delIds, delFromId)
		}
		if c.ToId == id {
			delToId, err := b.DeleteConnection(c.Id)
			if err != nil {
				return nil, errors.New(fmt.Sprintf("Cannot delete block %s: ToId %s does not exist", id, c.ToId))
			}
			delIds = append(delIds, delToId)
		}
	}

	// turn off block here
	// close channels, whatever.
	b.blockMap[id].chans.QuitChan <- true

	delete(b.blockMap, id)
	delIds = append(delIds, id)

	return delIds, nil
}

func (b *BlockManager) DeleteConnection(id string) (string, error) {
	_, ok := b.connMap[id]
	if !ok {
		return "", errors.New(fmt.Sprintf("Cannot delete connection %s: does not exist", id))
	}

	b.blockMap[b.connMap[id].FromId].chans.DelChan <- &blocks.Msg{
		Route: id,
	}

	b.connMap[id].chans.QuitChan <- true

	// call disconnecting stuff here
	// remove channel from FromBlock, etc
	// turn off connection block
	delete(b.connMap, id)

	return id, nil
}

func (b *BlockManager) StatusBlocks() []string {
	var wg sync.WaitGroup
	MsgChan := make(chan string, len(b.blockMap))
	for k, _ := range b.blockMap {
		wg.Add(1)
		go func(queryChan chan *blocks.QueryMsg) {
			defer wg.Done()
			timeout := time.NewTimer(time.Second * 5)
			var returnToSender blocks.MsgChan
			returnToSender = make(chan interface{})
			queryChan <- &blocks.QueryMsg{
				Route:   "ping",
				MsgChan: returnToSender,
			}
			select {
			case q := <-returnToSender:
				MsgChan <- q.(string)
			case <-timeout.C:
				MsgChan <- "TIMEOUT"
			}
		}(b.blockMap[k].chans.QueryChan)
	}
	wg.Wait()
	responses := make([]string, len(b.blockMap))
	for i := 0; i < len(b.blockMap); i++ {
		responses[i] = <-MsgChan
	}
	return responses
}

func (b *BlockManager) UpdateBlockId(fromId string, toId string) (*BlockInfo, []*ConnectionInfo, error) {
	_, ok := b.blockMap[fromId]
	if !ok {
		return nil, nil, errors.New("from block Id does not exist")
	}

	if !b.IdSafe(toId) {
		return nil, nil, errors.New(fmt.Sprintf("Cannot create block %s: invalid id", toId))
	}

	// make sure ID doesn't already exist
	if b.IdExists(toId) {
		return nil, nil, errors.New(fmt.Sprintf("Cannot create block %s: id already exists", toId))
	}

	select {
	case b.blockMap[fromId].chans.IdChan <- toId:
	default:
		return nil, nil, errors.New(fmt.Sprintf("Could not set Id for block %s: timeout", fromId))
	}

	b.blockMap[toId] = b.blockMap[fromId]
	b.blockMap[toId].Id = toId

	delete(b.blockMap, fromId)

	var updatedConns []*ConnectionInfo

	for _, c := range b.connMap {
		if c.FromId == fromId || c.ToId == fromId {
			updatedConns = append(updatedConns, c)
			if c.FromId == fromId {
				c.FromId = toId
			}
			if c.ToId == fromId {
				c.ToId = toId
			}
		}
	}

	return b.blockMap[toId], updatedConns, nil
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
