package blocks

import(
	"time"
)

type Msg struct {
	Msg   interface{}
	Route string
}

type AddChanMsg struct {
	Route   string
	Channel chan *Msg
}

type QueryMsg struct {
	Route    string
	RespChan chan interface{}
}

type BlockChans struct {
	InChan    chan *Msg
	QueryChan chan *QueryMsg
	AddChan   chan *AddChanMsg
	DelChan   chan *Msg
	ErrChan   chan error
	QuitChan  chan bool
}

type Block struct {
	Name        string // the name of the block specifed by the user (like MyBlock)
	Kind        string // the kind of block this is (like count, toFile, fromSQS)
	inRoutes    map[string]chan interface{}
	queryRoutes map[string]chan chan interface{}
	broadcast   chan interface{}
	quit 		chan interface{}
	doesBroadcast bool
	BlockChans
}

type BlockDef struct {
	Type string
	InRoutes []string
	QueryRoutes [] string
	OutRoutes []string
}

type BlockInterface interface {
	Setup()
	Run()
	CleanUp()
	Error(error)
	Build(BlockChans)
	Quit() chan interface{}
	Broadcast() chan interface{}
	InRoute(string) chan interface{}
	QueryRoute(string) chan chan interface{}
	GetBlock() *Block
	GetDef() *BlockDef
}

func (b *Block) Build(c BlockChans) {
	// fuck can I do this all in one?
	b.InChan = c.InChan
	b.QueryChan = c.QueryChan
	b.AddChan = c.AddChan
	b.DelChan = c.DelChan
	b.ErrChan = c.ErrChan
	b.QuitChan = c.QuitChan

	// route maps
	b.inRoutes = make(map[string]chan interface{})
	b.queryRoutes = make(map[string]chan chan interface{})

	// broadcast channel
	b.broadcast = make(chan interface{})

	// quit chan
	b.quit = make(chan interface{})
}

func (b *Block) InRoute(routeName string) chan interface{} {
	route := make(chan interface{})
	b.inRoutes[routeName] = route
	return route
}

func (b *Block) QueryRoute(routeName string) chan chan interface{} {
	route := make(chan chan interface{})
	b.queryRoutes[routeName] = route
	return route
}

func (b *Block) Broadcast() chan interface{} {
	b.doesBroadcast = true
	return b.broadcast
}

func (b *Block) Quit() chan interface{} {
	return b.quit
}

func (b *Block) GetBlock() *Block {
	return b
}

func (b *Block) GetDef() *BlockDef {
	var inRoutes []string
	var queryRoutes []string
	var outRoutes []string

	for k, _ := range b.inRoutes {
		inRoutes = append(inRoutes, k)
	}

	for k, _ := range b.queryRoutes {
		queryRoutes = append(queryRoutes, k)
	}

	if b.doesBroadcast {
		outRoutes = []string{"out"}
	}

	return &BlockDef{
		Type: b.Kind,
		InRoutes: inRoutes,
		QueryRoutes: queryRoutes, 
		OutRoutes: outRoutes,
	}
}

func (b *Block) CleanUp() {
	b.inRoutes["quit"] <- true
	for route := range b.inRoutes {
		defer close(b.inRoutes[route])
	}
	for route := range b.queryRoutes {
		defer close(b.queryRoutes[route])
	}
	defer close(b.InChan)
	defer close(b.QueryChan)
	defer close(b.AddChan)
	defer close(b.DelChan)
	defer close(b.ErrChan)
	defer close(b.QuitChan)
	defer close(b.broadcast)
}

func (b *Block) Error(e error) {
	b.ErrChan <- e
}

func BlockRoutine(bi BlockInterface) {
	outChans := make(map[string]chan *Msg)

	b := bi.GetBlock()
	bi.Setup()
	go bi.Run()

	for {
		select {
		case msg := <-b.InChan:
			b.inRoutes[msg.Route] <- msg.Msg
		case msg := <-b.QueryChan:
			b.queryRoutes[msg.Route] <- msg.RespChan
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
	InChan chan *Msg
	QueryChan chan *QueryMsg
	AddChan chan *AddChanMsg
	DelChan chan *Msg
	QuitChan chan bool
	ToRoute string
}

func ConnectionRoutine(c *Connection){
	var last interface{}
	var rate float64

	outChans := make(map[string]chan *Msg)
	times := make([]int64,100,100)
	timesIdx := len(times)

	for{
		select{
		case msg := <- c.InChan:
			last = msg.Msg

			for _, v := range outChans {
				v <- &Msg{
					Msg:   msg,
					Route: c.ToRoute,
				}
			}

			times = times[1:]
			times = append(times, time.Now().UnixNano())

			if timesIdx > 0 {
				timesIdx--
			}

		case msg := <- c.QueryChan:
			switch msg.Route {
			case "rate":
				if timesIdx == len(times) {
					rate = 0
				} else {
					rate = 1000000000.0 * float64(len(times) - timesIdx)/float64(time.Now().UnixNano() - times[timesIdx])
				}
				msg.RespChan <- rate
			case "last":
				msg.RespChan <- last
			}
		case msg := <- c.AddChan:
			outChans[msg.Route] = msg.Channel
		case msg := <- c.DelChan:
			delete(outChans, msg.Route)
		case <- c.QuitChan:
			return
		}
	}
}
