package blocks

import (
	"time"
)

// Connection accepts the input from a block and outputs it to another block.
// This block is a special case in that it requires an input and an output block
// to be created.
func Connection(b *Block) {
	var last BMsg
	var rate float64 // rate in messages per second of this block
	var N float64    // number of messages passed through this block
	var t time.Time
	for {
		select {
		case msg := <-b.InChan:
			last = msg
			broadcast(b.OutChans, msg)
			// rate calc
			if t.IsZero() {
				// this is the connection's first message
				t = time.Now()
				break
			}
			N++
			dt := time.Since(t).Seconds()
			rate = ((N-1.0)/N)*rate + (1.0/N)*dt
			t = time.Now()
		case query := <-b.Routes["last_message"]:
			rr, ok := query.(RouteResponse)
			if ok {
				rr.ResponseChan <- last
			}
		case query := <-b.Routes["rate"]:
			rr, ok := query.(RouteResponse)
			if ok {
				rr.ResponseChan <- map[string]float64{"rate": rate}
			}
		case msg := <-b.AddChan:
			updateOutChans(msg, b)
		case <-b.QuitChan:
			quit(b)
			return
		}
	}

}
