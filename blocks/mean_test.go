package blocks

import (
	"encoding/json"
	"testing"
)

func TestSetRule(t *testing.T) {
	msg := map[string]interface{}{
		"Key": "TestKey",
	}
	m, _ := json.Marshal(msg)
	err := TestRoute("avg", m, "set_rule")
	if err != nil {
		t.Error(err.Error())
	}
}

func TestGetRule(t *testing.T) {
	msg := map[string]interface{}{}
	m, _ := json.Marshal(msg)
	err := TestRoute("avg", m, "get_rule")
	if err != nil {
		t.Error(err.Error())
	}
}

func TestAvg(t *testing.T) {
	inChan := make(chan BMsg)
	successChan := make(chan bool)
	e, _ := json.Marshal(map[string]interface{}{"Avg": 100})
	r, _ := json.Marshal(map[string]interface{}{"Key": "value"})
	go TestState("avg", inChan, 5, e, r, "avg", successChan)
	inChan <- map[string]interface{}{
		"value": 150,
	}
	inChan <- map[string]interface{}{
		"value": 50,
	}
	if !<-successChan {
		t.Error("state did not match")
	}
}
