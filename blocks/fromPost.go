package blocks

// accepts JSON through POSTs to the /in endpoint and broadcasts to other blocks.
func FromPost(b *Block) {
	for {
		select {
		case msg := <-b.AddChan:
			updateOutChans(msg, b)
		case msg := <-b.Routes["in"]:
			marshal(msg, map[string]string{"POST": "OK"})
			out := BMsg{
				Msg:          msg.Msg,
				ResponseChan: nil,
			}
			broadcast(b.OutChans, &out)
		case <-b.QuitChan:
			quit(b)
			return
		}
	}
}
