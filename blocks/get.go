package blocks

// block designed to get values from other blocks
func GetRoute(b *Block) {

	type getRule struct {
	}
	var rule *getRule

	// this is the response channel for sending to ather blocks
	resChan := make(chan interface{})

	for {
		select {
		case m := <-b.Routes["set_rule"]:
			rule = &getRule{}
			unmarshal(m, rule)
		case r := <-b.Routes["get_rule"]:
			marshal(r, &getRule{})
		case <-b.InChan:
			// recieves a bang
			// queries its 'get' route
			req := BMsg{
				ResponseChan: resChan,
			}
			b.Routes["get"] <- req
			// the response will be picked up in another case
		case res := <-resChan:
			// waits for the response
			// and broadcasts it
			outMsg := BMsg{
				Msg: res,
			}
			broadcast(b.OutChans, outMsg)
		case msg := <-b.AddChan:
			updateOutChans(msg, b)
		case <-b.QuitChan:
			quit(b)
			return
		}
	}
}
