package util

import "errors"

func ParseBool(rule map[string]interface{}, key string) (bool, error) {
	var val bool
	var ok bool

	ruleI, ok := rule[key]
	if !ok {
		return val, errors.New("Key was not in rule")
	}
	val, ok = ruleI.(bool)
	if !ok {
		return val, errors.New("Key's value was not a bool")
	}
	return val, nil
}

func ParseString(rule map[string]interface{}, key string) (string, error) {
	var val string
	var ok bool

	ruleI, ok := rule[key]
	if !ok {
		return val, errors.New("Key was not in rule")
	}
	val, ok = ruleI.(string)
	if !ok {
		return val, errors.New("Key was not a string")
	}
	return val, nil
}

func ParseFloat(rule map[string]interface{}, key string) (float64, error) {
	var val float64
	var ok bool

	ruleI, ok := rule[key]
	if !ok {
		return val, errors.New("Key was not in rule")
	}
	val, ok = ruleI.(float64)
	if !ok {
		return val, errors.New("Key was not a float64")
	}
	return val, nil
}

func ParseInt(rule map[string]interface{}, key string) (int, error) {
	var val int
	var ok bool

	ruleI, ok := rule[key]
	if !ok {
		return val, errors.New("Key was not in rule")
	}
	val, ok = ruleI.(int)
	if !ok {
		return val, errors.New("Key was not an int!")
	}
	return val, nil
}

func CheckRule(messageI interface{}, ruleMsg map[string]string) bool {
	message := messageI.(map[string]string)
	for key, value := range ruleMsg {
		if message[key] != value {
			return false
		}
	}
	return true
}
