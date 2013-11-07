package blocks

import "time"

func Blocked(b *Block) {
	for {
		select {
		case <-b.InChan:
			break
		case msg := <-b.AddChan:
			updateOutChans(msg, b)
		case <-b.Routes["get_rule"]:
			c := time.NewTimer(time.Duration(6) * time.Second)
			<-c.C
		case <-b.QuitChan:
			quit(b)
			return
		}
	}
}
