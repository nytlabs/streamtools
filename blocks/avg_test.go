package blocks

import (
	"encoding/json"
	"log"
	"testing"
)

func TestCreate(t *testing.T) {
	BuildLibrary()
	b, err := NewBlock("avg", "testBlock")
	if err != nil {
		t.Error("failed to create avg block", err.Error())
	}
	go Library["avg"].Routine(b)
}

func TestSend(t *testing.T) {
	BuildLibrary()
	b, err := NewBlock("avg", "testBlock")
	if err != nil {
		t.Error("failed to create avg block", err.Error())
	}
	go Library["avg"].Routine(b)
	msg := make(map[string]interface{})
	msg["value"] = 2
	m, err := json.Marshal(msg)
	if err != nil {
		log.Fatal(err.Error())
	}
	b.InChan <- m

}
