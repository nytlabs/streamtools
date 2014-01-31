package blocks

func SkeletonTransfer(b *Block) {

	type skeletonRule struct {
		Param int
	}
	var rule *skeletonRule

	for {
		select {
		case m := <-b.Routes["set_rule"]:
			if rule == nil {
				rule = &skeletonRule{}
			}
			unmarshal(m, rule)
		case r := <-b.Routes["get_rule"]:
			if rule == nil {
				marshal(r, &skeletonRule{})
			} else {
				marshal(r, rule)
			}
		case msg := <-b.InChan:
			if rule == nil {
				break
			}
			broadcast(b.OutChans, msg)
		case msg := <-b.AddChan:
			updateOutChans(msg, b)
		case <-b.QuitChan:
			quit(b)
			return
		}
	}
}
