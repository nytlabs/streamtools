package blocks

import (
	"errors"
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
		//log.Println(string(<-rr.ResponseChan))
		successChan <- true
	}()
	select {
	case <-successChan:
	case <-timer.C:
		return errors.New("send failed")
	}
	return nil
}

/*
func TestBlocks(t *testing.T) {
	BuildLibrary()
	for _, b := range Library {
		log.Println("testing", b.BlockType)
		test_create(b.BlockType, t)
		test_send(b.BlockType, t)
		for _, r := range b.RouteNames {
			test_route(b.BlockType, r, t)
		}
	}
}

func create(b string) (*Block, error) {
	block, err := NewBlock(b, "testBlock")
	return block, err
}

func test_create(b string, t *testing.T) {
	block, err := create(b)
	if err != nil {
		t.Error("failed to create", b)
	}
	go Library[b].Routine(block)
}

func send(c chan BMsg, m BMsg, o chan bool) {
	c <- m
	o <- true
}

func timedSend(m BMsg, c chan BMsg) error {
	timer := time.NewTimer(time.Duration(5) * time.Second)
	responseChan := make(chan bool)
	go send(c, m, responseChan)
	select {
	case <-responseChan:
	case <-timer.C:
		return errors.New("send failed")
	}
	return nil

}

func timedSendRoute(m BMsg, c chan RouteResponse) error {
	timer := time.NewTimer(time.Duration(5) * time.Second)
	responseChan := make(chan bool)
	o, err := json.Marshal(m)
	if err != nil {
		log.Fatal(err.Error())
	}
	rr := RouteResponse{
		Msg:          o,
		ResponseChan: make(chan []byte),
	}
	go func() {
		c <- rr
		responseChan <- true
	}()
	select {
	case <-responseChan:
	case <-timer.C:
		return errors.New("send failed")
	}
	return nil
}

func test_send(b string, t *testing.T) {
	block, err := create(b)
	if err != nil {
		t.Error("failed to create", b)
	}
	go Library[b].Routine(block)
	msg := make(map[string]interface{})
	msg["value"] = 2
	m, err := json.Marshal(msg)
	if err != nil {
		log.Fatal(err.Error())
	}
	timedSend(m, block.InChan)
}

func test_route(b string, r string, t *testing.T) {
	block, err := create(b)
	if err != nil {
		t.Error("failed to create", b)
	}
	go Library[b].Routine(block)
	msg := make(map[string]interface{})
	m, err := json.Marshal(msg)
	if err != nil {
		log.Fatal(err.Error())
	}
	timedSendRoute(m, block.Routes[r])
	return err
}
*/
