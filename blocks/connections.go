package blocks

import (
	"time"
)

// Connection accepts the input from a block and outputs it to another block.
// This block is a special case in that it requires an input and an output block
// to be created.
func Connection(b *Block) {
	var last interface{}
	var rate float64 // rate in messages per second of this block

	times := make([]int64,100,100)
	timesIdx := len(times)

	for {
		select {
		case msg := <-b.InChan:
			last = msg.Msg
			broadcast(b.OutChans, msg)

			times = times[1:]
			times = append(times, time.Now().UnixNano())

			if timesIdx > 0 {
				timesIdx--
			}

		case query := <-b.Routes["last_message"]:
			query.ResponseChan <- last
		case query := <-b.Routes["rate"]:
			if timesIdx == len(times) {
				rate = 0
			} else {
				rate = 1000000000.0 * float64(len(times) - timesIdx)/float64(time.Now().UnixNano() - times[timesIdx])
			}

			query.ResponseChan <- map[string]float64{"rate": rate}
		case msg := <-b.AddChan:
			updateOutChans(msg, b)
		case <-b.QuitChan:
			quit(b)
			return
		}
	}

}
