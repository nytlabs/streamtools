package blocks

func Blocked(b *Block) {
	for {
		select {
		case <-b.InChan:
			break
		case msg := <-b.AddChan:
			updateOutChans(msg, b)
		case <-b.Routes["get_rule"]:
		case <-b.QuitChan:
			quit(b)
			return
		}
	}
}
