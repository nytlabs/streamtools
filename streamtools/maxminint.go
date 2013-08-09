package streamtools

import (
	"github.com/bitly/go-simplejson"
	"log"
	"math"
)

func maxInSlice(s interface{}) float64 {
	max := math.Inf(-1)
	switch s.(type) {
	case []float64:
		s, _ := s.([]float64)
		for _, ss := range s {
			max = math.Max(ss, max)
		}
	case []float32:
		s, _ := s.([]float32)
		for _, ss := range s {
			max = math.Max(float64(ss), max)
		}
	case []int:
		s, _ := s.([]int)
		for _, ss := range s {
			max = math.Max(float64(ss), max)
		}
	default:
		s, ok := s.([]interface{})
		if ok {
			for _, ss := range s {
				switch ss := ss.(type) {
				case int:
					max = math.Max(float64(ss), max)
				case float32:
					max = math.Max(float64(ss), max)
				case float64:
					max = math.Max(float64(ss), max)
				}
			}
		} else {
			return math.NaN()
		}
	}
	return max
}

var TrackMax STTrackingFunc = func(inChan chan simplejson.Json, route string, port int) {
	store := NewStore()
	go store.serveHTTP(route, port)
	for {
		select {
		case m := <-inChan:
			blob, err := m.Map()
			if err != nil {
				log.Fatalln(err)
			}
			log.Println(blob)
			store.storeMax(blob, nil)
		}
	}
}
