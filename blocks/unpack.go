package blocks

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
			array := getKeyValues(msg, rule.Path)
			for _, outMsg := range array {
				broadcast(b.OutChans, outMsg)
			}
		case msg := <-b.AddChan:
			updateOutChans(msg, b)
		case <-b.QuitChan:
			quit(b)
			return
		}
	}
}
