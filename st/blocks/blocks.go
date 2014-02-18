package blocks

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
	BlockChans
}

type BlockInterface interface {
	Setup()
	Run()
	Quit()
	Build(BlockChans)
	Broadcast() chan interface{}
	InRoute(string) chan interface{}
	QueryRoute(string) chan chan interface{}
	GetBlock() *Block
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
	return b.broadcast
}

func (b *Block) GetBlock() *Block {
	return b
}

func (b *Block) Quit() {
	b.inRoutes["quit"] <- true
	for route := range b.inRoutes {
		close(b.inRoutes[route])
	}
	for route := range b.queryRoutes {
		close(b.queryRoutes[route])
	}
	close(b.InChan)
	close(b.QueryChan)
	close(b.AddChan)
	close(b.DelChan)
	close(b.ErrChan)
	close(b.QuitChan)
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
			b.Quit()
			return
		}
	}
}
