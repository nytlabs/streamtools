package blocks

// PostTo accepts JSON through POSTs to the /in endpoint and broadcasts to other blocks.
func PostTo(b *Block) {
	for {
		select {
		case msg := <-b.AddChan:
			updateOutChans(msg, b)
		case msg := <-b.Routes["in"]:
			if rr, ok := msg.(RouteResponse); ok{
				marshal(msg, map[string]string{"POST":"OK",})
				broadcast(b.OutChans, rr.Msg)
			}
		case <-b.QuitChan:
			quit(b)
			return
		}
	}
}
