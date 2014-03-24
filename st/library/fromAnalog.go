package library

import (
	"github.com/mrmorphic/hwio"                // hwio
	"github.com/nytlabs/streamtools/st/blocks" // blocks
	"github.com/nytlabs/streamtools/st/util"
	"time"
)

// specify those channels we're going to use to communicate with streamtools
type FromAnalog struct {
	blocks.Block
	queryrule chan chan interface{}
	inrule    chan interface{}
	out       chan interface{}
	quit      chan interface{}
}

// we need to build a simple factory so that streamtools can make new blocks of this kind
func NewFromAnalog() blocks.BlockInterface {
	return &FromAnalog{}
}

// Setup is called once before running the block. We build up the channels and specify what kind of block this is.
func (b *FromAnalog) Setup() {
	b.Kind = "FromAnalog"
	b.inrule = b.InRoute("rule")
	b.queryrule = b.QueryRoute("rule")
	b.quit = b.Quit()
	b.out = b.Broadcast()
}

// Run is the block's main loop. Here we listen on the different channels we set up.
func (b *FromAnalog) Run() {
	interval := time.Duration(1) * time.Second
	var pin hwio.Pin
	var pinStr, intervalStr string
	var err error
	sampler := time.NewTicker(interval)
// Get the module
m, e := hwio.GetAnalogModule()
if e != nil {
b.Log(e)
}
// Enable it.
e = m.Enable()
if e != nil {
b.Log(e)
}
	for {
		select {
		case ruleI := <-b.inrule:
			// set a parameter of the block
			rule, ok := ruleI.(map[string]interface{})
			if !ok {
				b.Error("couldn't conver rule to map")
				continue
			}
			intervalStr, err = util.ParseString(rule, "Interval")
			if err != nil {
				b.Error(err)
				continue
			}
			interval, err = time.ParseDuration(intervalStr)
			if err != nil {
				b.Error(err)
				continue
			}
			pinStr, err = util.ParseString(rule, "Pin")
			if err != nil {
				b.Error(err)
				continue
			}
			pin, err = hwio.GetPin(pinStr)
			if err != nil {
				b.Error(err)
				continue
			}
			err = hwio.PinMode(pin, hwio.INPUT)
			if err != nil {
				b.Error(err)
				continue
			}
			// restart the sampler
			sampler = time.NewTicker(interval)
		case <-b.quit:
			// quit the block
			hwio.CloseAll() // TODO only close the pin associated with this block
			return
		case c := <-b.queryrule:
			// deal with a query request
			c <- map[string]interface{}{
				"Interval": intervalStr,
				"Pin":      pinStr,
			}

		case <-sampler.C:
			if pin == 0 {
				continue
			}
			v, err := hwio.AnalogRead(pin)
			if err != nil {
				b.Error(err)
				continue
			}
			out := map[string]interface{}{
				"value": v,
				"pin":   pinStr,
			}
			b.out <- out
		}
	}
}
