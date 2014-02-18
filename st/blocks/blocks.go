package blocks

import(
    "sync"

)

type Msg struct {
    msg interface{}
    route string
}

type Block struct {
    Name  string // the name of the block specifed by the user (like MyBlock)
    Kind  string // the kind of block this is (like count, toFile, fromSQS)
    inRoutes map[string]chan interface{}
    queryRoutes map[string]chan chan interface{}
    broadcast chan interface{}
    BlockChans
}

func (b *Block) Alloc(){
    b.inRoutes = make(map[string]chan interface{})
    b.queryRoutes = make(map[string]chan chan interface{})
    b.broadcast = make(chan interface{})

    b.InChan = make(chan *Msg)
    b.OutChans = make(map[string]chan *Msg)

    b.Mu = &sync.Mutex{}
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

func (b *Block) GetBlock() *Block{
    return b
}

type BlockChans struct {
    InChan chan *Msg
    QueryChan chan *Msg
    AddChan chan *Msg
    DelChan chan *Msg
    Mu      *sync.Mutex
}

type BlockInterface interface {
    Alloc()
    Setup()
    InRoute()   chan interface{}
    QueryRoute() chan chan interface{}
    Broadcast()
    Run()
    GetBlock() *Block
}

func blockRoutine(bi BlockInterface){
    OutChans  = make(map[string]chan *Msg)

    b := bi.GetBlock()

    bi.Setup()
    go bi.Run()

    for{
        select{
        case msg := <-b.InChan:
            b.inRoutes[msg.route] <- msg.msg
        case msg := <-b.QueryChan:
            b.queryRoutes[msg.route] <- msg.msg.(chan interface{})
        case msg := <-b.AddChan:
            OutChans[msg.route] = msg.msg.(chan *Msg)
        case msg := <-b.DelChan:
            delete(OutChans, msg.route)
        case msg := <- b.broadcast:
            for _, v := range OutChans{
                v <- &Msg{
                    msg: msg,
                    route: "",
                }
            }
        }
    }
}

