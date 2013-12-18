package blocks

// unpack an array into seperate messages, and emit them in order
func Unpack(b *Block) {

	type unpackRule struct {
		Path string
	}

	var rule *unpackRule

	for {
		select {
		case m := <-b.Routes["set_rule"]:
			if rule == nil {
				rule = &unpackRule{}
			}
			unmarshal(m, rule)
		case r := <-b.Routes["get_rule"]:
			if rule == nil {
				marshal(r, &unpackRule{})
			} else {
				marshal(r, rule)
			}
		case msg := <-b.InChan:
			if rule == nil {
				break
			}
			array := getKeyValues(msg.Msg, rule.Path)
			for _, outMsg := range array {
				out := BMsg{
					Msg:          outMsg,
					ResponseChan: nil,
				}
				broadcast(b.OutChans, out)
			}
		case msg := <-b.AddChan:
			updateOutChans(msg, b)
		case <-b.QuitChan:
			quit(b)
			return
		}
	}
}
