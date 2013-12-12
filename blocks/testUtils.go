package blocks

/*import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"time"
)

// make sure the route accepts the message
func TestRoute(blockType string, m []byte, route string) error {
	// build the library
	BuildLibrary()
	// create the block
	block, err := NewBlock(blockType, "testBlock")
	if err != nil {
		return err
	}
	// set the block running
	go Library[blockType].Routine(block)
	// set the timer running
	timer := time.NewTimer(time.Duration(5) * time.Second)
	// build the channel a success will be broadcast on
	successChan := make(chan bool)
	// build the route response
	rr := RouteResponse{
		Msg:          m,
		ResponseChan: make(chan []byte),
	}
	// send the test response
	go func() {
		block.Routes[route] <- rr
		<-rr.ResponseChan
		successChan <- true
	}()
	select {
	case <-successChan:
	case <-timer.C:
		return errors.New("send failed")
	}
	return nil
}

// make sure block has the expected state after a given time.
// send messages in on the inChan, and after the duration the state will be
// compared to the expected byte sequence.
// blockType : the kind of block to test
// inChan : send data through this channel to affect the state
// t : wait this many seconds before checking the state
// expected : compare the state to this expected byte array
// rule : set the block with this rule
// route : test the state using this route
func TestState(blockType string, inChan chan BMsg, t int, expected []byte, rule []byte, route string, successChan chan bool) {
	// build the library
	BuildLibrary()
	// create the block
	block, _ := NewBlock(blockType, "testBlock")
	// set the block running
	go Library[blockType].Routine(block)
	// set the rule
	rr := RouteResponse{
		Msg:          rule,
		ResponseChan: make(chan []byte),
	}
	block.Routes["set_rule"] <- rr
	<-rr.ResponseChan
	// set the timer running
	timer := time.NewTimer(time.Duration(t) * time.Second)
	for {
		select {
		case m := <-inChan:
			block.InChan <- m
		case <-timer.C:
			m, _ := json.Marshal("{}")
			rr := RouteResponse{
				Msg:          m,
				ResponseChan: make(chan []byte),
			}
			block.Routes[route] <- rr
			state := <-rr.ResponseChan
			if bytes.Equal(state, expected) {
				successChan <- true
			} else {
				log.Println("got", string(state), "expected", string(expected))
				successChan <- false
			}
			break
		}
	}
}*/
