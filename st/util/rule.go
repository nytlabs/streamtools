package util

import (
	"errors"
	"github.com/nytlabs/gojee"
)

func ParseBool(ruleI interface{}, key string) (bool, error) {
	rule := ruleI.(map[string]interface{})
	var val bool
	var ok bool

	foundRule, ok := rule[key]
	if !ok {
		return val, errors.New("Key was not in rule")
	}
	val, ok = foundRule.(bool)
	if !ok {
		return val, errors.New("Key's value was not a bool")
	}
	return val, nil
}

func ParseString(ruleI interface{}, key string) (string, error) {
	rule := ruleI.(map[string]interface{})
	var val string
	var ok bool

	foundRule, ok := rule[key]
	if !ok {
		return val, errors.New("Key was not in rule")
	}
	val, ok = foundRule.(string)
	if !ok {
		return val, errors.New("Key was not a string")
	}
	return val, nil
}

func ParseFloat(ruleI interface{}, key string) (float64, error) {
	rule := ruleI.(map[string]interface{})
	var val float64
	var ok bool

	foundRule, ok := rule[key]
	if !ok {
		return val, errors.New("Key was not in rule")
	}
	val, ok = foundRule.(float64)
	if !ok {
		return val, errors.New("Key was not a float64")
	}
	return val, nil
}

func ParseInt(ruleI interface{}, key string) (int, error) {
	rule := ruleI.(map[string]interface{})
	var val int
    var ok bool
    var floatval float64
	foundRule, ok := rule[key]
	if !ok {
		return val, errors.New("Key was not in rule")
	}
    floatval, ok = foundRule.(float64)
    if !ok {
        return val, errors.New("Key was not a number")
    }
	val = int(floatval)
	return val, nil
}

func BuildTokenTree(path string) (tree *jee.TokenTree, err error) {
	token, err := jee.Lexer(path)
	if err != nil {
		return nil, err
	}
	return jee.Parser(token)
}
