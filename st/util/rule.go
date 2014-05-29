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

func ParseRequiredString(ruleI interface{}, key string) (string, error) {
	val, err := ParseString(ruleI, key)
	if err != nil {
		return val, err
	}
	if len(val) == 0 {
		return val, errors.New(key + " was an empty string")
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

func KeyExists(ruleI interface{}, key string) bool {
	rule := ruleI.(map[string]interface{})
	_, ok := rule[key]
	return ok
}

func ParseArrayString(ruleI interface{}, key string) ([]string, error) {
	var val []string

	rule := ruleI.(map[string]interface{})
	var ok bool
	foundRule, ok := rule[key]
	if !ok {
		return val, errors.New("Path was not in rule")
	}

	switch foundRule.(type) {
	case []interface{}:
		valI, ok := foundRule.([]interface{})
		if !ok {
			return val, errors.New("Supplied value was not an array of interfaces")
		}
		val = make([]string, len(valI))
		for i, vi := range valI {
			v, ok := vi.(string)
			if !ok {
				return val, errors.New("Failed asserting to []string")
			}
			val[i] = v
		}
	case []string:
		val, ok = foundRule.([]string)
		if !ok {
			return val, errors.New("Supplied value was not an array of strings")
		}
	}
	return val, nil
}

func ParseArrayFloat(ruleI interface{}, key string) ([]float64, error) {
	rule := ruleI.(map[string]interface{})
	var ok bool
	var val []float64
	foundRule, ok := rule[key]
	if !ok {
		return val, errors.New("Path was not in rule")
	}
	valI, ok := foundRule.([]interface{})
	if !ok {
		return val, errors.New("Supplied value was not an array")
	}
	val = make([]float64, len(valI))
	for i, vi := range valI {
		v, ok := vi.(float64)
		if !ok {
			return val, errors.New("Supplied value was not an array of numbers")
		}
		val[i] = v
	}
	return val, nil
}

func BuildTokenTree(path string) (tree *jee.TokenTree, err error) {
	token, err := jee.Lexer(path)
	if err != nil {
		return nil, err
	}
	return jee.Parser(token)
}
