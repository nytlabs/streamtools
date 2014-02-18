package blocks

type Msg struct {
	msg   interface{}
	route string
}

type AddChanMsg struct {
	route   string
	channel chan *Msg
}

type QueryMsg struct {
	route    string
	respChan chan interface{}
}

type BlockChans struct {
	InChan    chan *Msg
	QueryChan chan *QueryMsg
	AddChan   chan *AddChanMsg
	DelChan   chan *Msg
	ErrChan   chan error
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

func BlockRoutine(bi BlockInterface) {
	outChans := make(map[string]chan *Msg)

	b := bi.GetBlock()

	bi.Setup()
	go bi.Run()

	for {
		select {
		case msg := <-b.InChan:
			b.inRoutes[msg.route] <- msg.msg
		case msg := <-b.QueryChan:
			b.queryRoutes[msg.route] <- msg.respChan
		case msg := <-b.AddChan:
			outChans[msg.route] = msg.channel
		case msg := <-b.DelChan:
			delete(outChans, msg.route)
		case msg := <-b.broadcast:
			for _, v := range outChans {
				v <- &Msg{
					msg:   msg,
					route: "",
				}
			}
		}
	}
}
