package streamtools

import (
	"fmt"
	"github.com/bitly/go-simplejson"
	"log"
	"math"
	"net/http"
	"strconv"
)

type Store struct {
	data map[string]interface{}
}

// NewStore initializes and returns a new Store instance.
func NewStore() *Store {
	return &Store{
		data: make(map[string]interface{}),
	}
}

// convertMapToJson simply takes a map of strings to strings,
// and converts it to a simplejson.Json object.
func convertMapToJson(m map[string]interface{}) *simplejson.Json {
	msg, _ := simplejson.NewJson([]byte("{}"))
	for k, v := range m {
		msg.Set(k, v)
	}
	return msg
}

func (self Store) prettyPrintData() string {
	return "dummy"
}

func (self Store) httpHandler(w http.ResponseWriter, r *http.Request) {
	//	fmt.Fprintf(w, "Store sez hi %s!", self.data)
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, self.data)
	//	fmt.Fprint(w, self.prettyPrintData())
	// for k, _ := range self.data {
	// 	fmt.Fprintf(w, k)
	// }
}

func (self Store) serveHTTP(route string, port int) {
	http.HandleFunc(route, self.httpHandler)
	http.ListenAndServe(":"+strconv.Itoa(port), nil)
	log.Printf("Serving HTTP \"%s\" on port %v", route, port)
}

//////////////////

func (self Store) storeIsIntProbability(m simplejson.Json) {
}

func (self Store) storeMin(m simplejson.Json) {
}

func (self Store) storeMax(d map[string]interface{}, s map[string]interface{}) {
	if s == nil {
		s = self.data
	}
	for k, v := range d {
		switch v := v.(type) {
		case map[string]interface{}:
			// got an object
			if _, ok := s[k]; !ok {
				s[k] = make(map[string]interface{})
			}
			vv, _ := s[k].(map[string]interface{})
			self.storeMax(v, vv)
		case []interface{}:
			// got an array
			m := maxInSlice(v)
			log.Printf("key=%s: array got max=%v", k, m)
			if !math.IsNaN(m) {
				vv, _ := s[k].(float64)
				s[k] = math.Max(m, vv)
			}
		case int, float32, float64:
			// got a number
			if s[k] == nil {
				s[k] = math.Inf(-1)
			}
			vv, _ := s[k].(float64)
			switch v := v.(type) {
			case int:
				s[k] = math.Max(float64(v), vv)
			case float32:
				s[k] = math.Max(float64(v), vv)
			case float64:
				s[k] = math.Max(v, vv)
			}
		default:
			// nil, string, bool; do nothing.
		}
	}
}
