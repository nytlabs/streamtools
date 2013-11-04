package blocks

// this is a skeleton state block. It doesn't do anything, but can be used as a
// template to make new state blocks.
func SkeletonState(b *Block) {

	// the rule defines how the block works. It can be get and set live
	type skeletonRule struct {
		Param int
	}

	// the state defines the current state of the block. This is where we store
	// whatever is being learned about the data
	type skeletonState struct {
		Data int
	}

	// we initialise the data and the rule. It's often helpful to provide an
	// initial state of the block
	data := &skeletonState{Data: 0}
	var rule *skeletonRule

	for {
		select {
		case query := <-b.Routes["stateRoute"]:
			// the state route is used to retrieve the state of the block
			marshal(query, data)
		case ruleUpdate := <-b.Routes["set_rule"]:
			// the set rule route is used to update the block's rule. There are two
			// situations: if we've not seen any rules before - i.e. this is a brand
			// new block, or if this is a block that already had rules.
			if rule == nil {
				rule = &skeletonRule{}
			}
			unmarshal(ruleUpdate, rule)
		case msg := <-b.Routes["get_rule"]:
			// the get rule route reports the current rule being used by the block
			if rule == nil {
				marshal(msg, &skeletonRule{})
			} else {
				marshal(msg, rule)
			}
		case <-b.QuitChan:
			// the quit chan recieves a signal that tells this block to
			// terminate. Most of the time all you need do is pass the block to
			// the quit function
			quit(b)
			return
		case <-b.InChan:
			// the in chan recieves in bound messages. This is usually where the
			// bulk of the processing goes. In a state block, this is where the
			// state should be updated
			if rule == nil {
				break
			}
			data.Data++
		}
	}
}
