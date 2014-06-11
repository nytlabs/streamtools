package blocks

import (
	"fmt"
	"github.com/nytlabs/streamtools/st/loghub"
	"net/url"
	"strconv"
	"time"
)

type Msg struct {
	Msg   interface{}
	Route string
}

type MsgChan chan interface{}

func (c MsgChan) MarshalJSON() ([]byte, error) {
	outBytes := []byte("{\"channel\":" + strconv.Itoa(len(c)) + "}")
	return outBytes, nil
}

type AddChanMsg struct {
	Route   string
	Channel chan *Msg
}

type QueryMsg struct {
	Route   string
	MsgChan MsgChan
}

type QueryParamMsg struct {
	Route    string
	RespChan chan interface{}
	Params   url.Values
}

type Query struct {
	RespChan chan interface{}
	Params   url.Values
}

type BlockChans struct {
	InChan         chan *Msg
	QueryChan      chan *QueryMsg
	QueryParamChan chan *QueryParamMsg
	AddChan        chan *AddChanMsg
	DelChan        chan *Msg
	IdChan         chan string
	ErrChan        chan error
	QuitChan       chan bool
}

type LogStreams struct {
	log MsgChan
	ui  MsgChan
}

type Block struct {
	Id               string // the name of the block specifed by the user (like MyBlock)
	Kind             string // the kind of block this is (like count, toFile, fromSQS)
	Desc             string // the description of block ('counts the number of messages it has seen')
	inRoutes         map[string]MsgChan
	queryRoutes      map[string]chan MsgChan
	queryParamRoutes map[string]chan Query
	broadcast        MsgChan
	quit             MsgChan
	doesBroadcast    bool
	BlockChans
	LogStreams
}

type BlockDef struct {
	Type             string
	Desc             string
	InRoutes         []string
	QueryRoutes      []string
	QueryParamRoutes []string
	OutRoutes        []string
}

type BlockInterface interface {
	Setup()
	Run()
	CleanUp()
	Build(BlockChans)
	Quit() MsgChan
	Broadcast() MsgChan
	InRoute(string) MsgChan
	QueryRoute(string) chan MsgChan
	QueryParamRoute(string) chan Query
	GetBlock() *Block
	GetDef() *BlockDef
	Log(interface{})
	Error(interface{})
	SetId(string)
}

func (b *Block) Build(c BlockChans) {
	// block channels
	b.InChan = c.InChan
	b.QueryChan = c.QueryChan
	b.QueryParamChan = c.QueryParamChan
	b.AddChan = c.AddChan
	b.DelChan = c.DelChan
	b.ErrChan = c.ErrChan
	b.QuitChan = c.QuitChan
	b.IdChan = c.IdChan

	// route maps
	b.inRoutes = make(map[string]MsgChan) // necessary to stop locking...
	b.queryRoutes = make(map[string]chan MsgChan)
	b.queryParamRoutes = make(map[string]chan Query)

	// broadcast channel
	b.broadcast = make(MsgChan, 10) // necessary to stop locking...

	// quit chan
	b.quit = make(MsgChan)

	b.ui = make(MsgChan)
	b.log = make(MsgChan)
}

func (b *Block) SetId(Id string) {
	b.Id = Id
}

func (b *Block) InRoute(routeName string) MsgChan {
	route := make(MsgChan, 1000)
	b.inRoutes[routeName] = route
	return route
}

func (b *Block) QueryRoute(routeName string) chan MsgChan {
	route := make(chan MsgChan, 1000)
	b.queryRoutes[routeName] = route
	return route
}

func (b *Block) QueryParamRoute(routeName string) chan Query {
	route := make(chan Query, 1000)
	b.queryParamRoutes[routeName] = route
	return route
}

func (b *Block) Broadcast() MsgChan {
	b.doesBroadcast = true
	return b.broadcast
}

func (b *Block) Quit() MsgChan {
	return b.quit
}

func (b *Block) GetBlock() *Block {
	return b
}

func (b *Block) GetDef() *BlockDef {
	inRoutes := []string{}
	queryRoutes := []string{}
	queryParamRoutes := []string{}
	outRoutes := []string{}

	for k, _ := range b.inRoutes {
		inRoutes = append(inRoutes, k)
	}

	for k, _ := range b.queryRoutes {
		queryRoutes = append(queryRoutes, k)
	}

	for k, _ := range b.queryParamRoutes {
		queryParamRoutes = append(queryParamRoutes, k)
	}

	if b.doesBroadcast {
		outRoutes = []string{"out"}
	}

	return &BlockDef{
		Type:             b.Kind,
		Desc:             b.Desc,
		InRoutes:         inRoutes,
		QueryRoutes:      queryRoutes,
		QueryParamRoutes: queryParamRoutes,
		OutRoutes:        outRoutes,
	}
}

func (b *Block) CleanUp() {
	for route := range b.inRoutes {
		defer close(b.inRoutes[route])
	}
	for route := range b.queryRoutes {
		defer close(b.queryRoutes[route])
	}
	defer close(b.InChan)
	defer close(b.QueryChan)
	defer close(b.QueryParamChan)
	defer close(b.AddChan)
	defer close(b.DelChan)
	defer close(b.ErrChan)
	defer close(b.QuitChan)
	defer close(b.broadcast)
	defer close(b.IdChan)

	go func(id string) {
		loghub.Log <- &loghub.LogMsg{
			Type: loghub.INFO,
			Data: fmt.Sprintf("Block %s Quitting...", b.Id),
			Id:   id,
		}
	}(b.Id)
}

func (b *Block) Error(msg interface{}) {
	go func(id string) {
		loghub.Log <- &loghub.LogMsg{
			Type: loghub.ERROR,
			Data: msg,
			Id:   id,
		}
	}(b.Id)
}

func (b *Block) Log(msg interface{}) {
	go func(id string) {
		loghub.Log <- &loghub.LogMsg{
			Type: loghub.INFO,
			Data: msg,
			Id:   id,
		}
	}(b.Id)
}

func BlockRoutine(bi BlockInterface) {
	var dropped int64
	dropTicker := time.NewTicker(time.Duration(1 * time.Second))
	dropTicker.Stop()

	outChans := make(map[string]chan *Msg)
	b := bi.GetBlock()
	bi.Setup()
	go bi.Run()

	for {
		select {
		case <-dropTicker.C:
			go func(id string, count int64) {
				loghub.Log <- &loghub.LogMsg{
					Type: loghub.ERROR,
					Data: fmt.Sprintf("Dropped messages: %d (cannot keep up with stream)", count),
					Id:   id,
				}
			}(b.Id, dropped)

			if dropped == 0 {
				dropTicker.Stop()
			}

			dropped = 0
		case msg := <-b.InChan:
			_, ok := b.inRoutes[msg.Route]
			if !ok {
				break
			}

			// every in channel is buffered a 1000 messages.
			// if we cannot immediately send to that in channel we place the msg
			// in a go routine and notify the user that the block routine's
			// buffer has overflowed. this still allows for unrecoverable
			// overflows (for example: a stuck run() function), but at least it
			// offloads to memory/cpu instead of blocking.
			select {
			case b.inRoutes[msg.Route] <- msg.Msg:
			default:
				if dropped == 0 {
					dropTicker.Stop()
					dropTicker = time.NewTicker(1 * time.Second)
				}

				dropped++
			}

			if msg.Route == "rule" {
				go func(id string) {
					loghub.UI <- &loghub.LogMsg{
						Type: loghub.RULE_UPDATED,
						Data: map[string]interface{}{},
						Id:   id,
					}
				}(b.Id)
			}

		case msg := <-b.QueryChan:

			if msg.Route == "ping" {
				msg.MsgChan <- "OK"
				continue
			}

			_, ok := b.queryRoutes[msg.Route]
			if !ok {
				break
			}

			select {
			case b.queryRoutes[msg.Route] <- msg.MsgChan:
			default:
				go func() {
					b.queryRoutes[msg.Route] <- msg.MsgChan
				}()
			}
		case msg := <-b.QueryParamChan:

			if msg.Route == "ping" {
				msg.RespChan <- "OK"
				continue
			}

			_, ok := b.queryParamRoutes[msg.Route]
			if !ok {
				break
			}
			q := Query{
				RespChan: msg.RespChan,
				Params:   msg.Params,
			}

			select {
			case b.queryParamRoutes[msg.Route] <- q:
			default:
				go func() {
					b.queryParamRoutes[msg.Route] <- q
				}()
			}
		case id := <-b.IdChan:
			b.SetId(id)
		case msg := <-b.AddChan:
			outChans[msg.Route] = msg.Channel
		case msg := <-b.DelChan:
			delete(outChans, msg.Route)
		case msg := <-b.broadcast:
			for _, v := range outChans {
				v <- &Msg{
					Msg:   msg,
					Route: "",
				}
			}
		case <-b.QuitChan:
			b.quit <- true
			b.CleanUp()
			return
		}
	}
}

type Connection struct {
	Id      string
	ToRoute string
	BlockChans
	LogStreams
}

func (c *Connection) SetId(Id string) {
	c.Id = Id
}

func (c *Connection) Build(chans BlockChans) {
	c.InChan = chans.InChan
	c.QueryChan = chans.QueryChan
	c.QueryParamChan = chans.QueryParamChan
	c.AddChan = chans.AddChan
	c.DelChan = chans.DelChan
	c.QuitChan = chans.QuitChan
}

func (c *Connection) CleanUp() {
	defer close(c.InChan)
	defer close(c.QueryChan)
	defer close(c.AddChan)
	defer close(c.DelChan)
	defer close(c.QuitChan)

	loghub.Log <- &loghub.LogMsg{
		Type: loghub.INFO,
		Data: fmt.Sprintf("Connection %s Quitting...", c.Id),
		Id:   c.Id,
	}
	close(c.QueryParamChan)
}

func ConnectionRoutine(c *Connection) {
	var last interface{}
	var rate float64

	outChans := make(map[string]chan *Msg)
	times := make([]int64, 100, 100)
	timesIdx := len(times)
	rateReport := time.NewTicker(200 * time.Millisecond)

	for {
		select {
		case <-rateReport.C:
			if timesIdx == len(times) {
				rate = 0
			} else {
				rate = 1000000000.0 * float64(len(times)-timesIdx) / float64(time.Now().UnixNano()-times[timesIdx])
			}

			go func(id string, r float64) {
				loghub.UI <- &loghub.LogMsg{
					Type: loghub.UPDATE_RATE,
					Data: map[string]interface{}{
						"Rate": r,
					},
					Id: id,
				}
			}(c.Id, rate)

		case msg := <-c.InChan:
			last = msg.Msg
			for _, v := range outChans {
				v <- &Msg{
					Msg:   msg.Msg,
					Route: c.ToRoute,
				}
			}

			times = times[1:]
			times = append(times, time.Now().UnixNano())

			if timesIdx > 0 {
				timesIdx--
			}

		case msg := <-c.QueryChan:
			switch msg.Route {
			case "rate":
				msg.MsgChan <- map[string]interface{}{
					"Rate": rate,
				}
			case "last":
				msg.MsgChan <- map[string]interface{}{
					"Last": last,
				}
			}
		case msg := <-c.AddChan:
			outChans[msg.Route] = msg.Channel
		case msg := <-c.DelChan:
			delete(outChans, msg.Route)
		case <-c.QuitChan:
			c.CleanUp()
			return
		}
	}
}
