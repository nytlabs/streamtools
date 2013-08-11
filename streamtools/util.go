package streamtools

import (
	"encoding/json"
	"github.com/bitly/go-simplejson"
	"log"
)

//// JSON Conversion ////

// TODO: This is really ugly. Ideally we have a function that takes and map[string]anything and converts it to simplejson (or whatever standardized json we want in the future.)  However, maps are mutable and therefore the set of map[string]interface{} types does not include map[string]string types. This is a temp fix.

// convertInterfaceMapToSimplejson simply takes a map of strings to
// interface{}, and converts it to a simplejson.Json object.
func convertInterfaceMapToSimplejson(m map[string]interface{}) *simplejson.Json {
	b, err := json.Marshal(m)
	if err != nil {
		log.Println("error:", err)
	}
	j, err := simplejson.NewJson(b)
	if err != nil {
		log.Println("error:", err)
	}
	return j
}

// convertStringMapToJson simply takes a map of strings to strings,
// and converts it to a simplejson.Json object.
func convertStringMapToSimplejson(m map[string]string) *simplejson.Json {
	msg, _ := simplejson.NewJson([]byte("{}"))
	for k, v := range m {
		msg.Set(k, v)
	}
	return msg
}

//// Set Math ////

// Don't know how we feel about external dependencies, but if there's
// demands for more comprehensive set math, we can consider just
// adopting this library wholesale: https://github.com/deckarep/golang-set

// N.B. instances of struct{} and [0]byte take 0 space in memory; see https://groups.google.com/d/msg/golang-nuts/lb4xLHq7wug/rArBbstQfSoJ

type Set map[interface{}]struct{}

// Creates and returns a reference to an empty set.
func NewSet() Set {
	return make(Set)
}

// NewSetFromSlice creates and returns a reference to a set from an existing slice.
func NewSetFromSlice(s []interface{}) Set {
	a := NewSet()
	for _, item := range s {
		a.Add(item)
	}
	return a
}

// Add adds an item to the set.
func (set Set) Add(i interface{}) bool {
	_, found := set[i]
	set[i] = struct{}{}
	return !found //returns false if element existed already
}

// Contains returns true if the specified item is in the set, false otherwise.
func (set Set) Contains(i interface{}) bool {
	_, found := set[i]
	return found
}

// Difference returns a new set with items in the current set but not in the other set.
func (set Set) Difference(other Set) Set {
	d := NewSet()
	for e := range set {
		if !other.Contains(e) {
			d.Add(e)
		}
	}
	return d
}

// ToSlice returns an []interface{} slice that represents the set.
func (set Set) ToSlice() []interface{} {
	r := []interface{}{}
	for e := range set {
		r = append(r, e)
	}
	return r
}
