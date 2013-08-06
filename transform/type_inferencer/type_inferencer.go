package main

import (
	"flag"
	"github.com/bitly/go-simplejson"
	"github.com/bitly/nsq/nsq"
	"log"
	"reflect"
)

var (
	inTopic          = flag.String("in_topic", "", "topic to read from")
	outTopic         = flag.String("out_topic", "", "topic to write to")
	lookupdHTTPAddrs = "127.0.0.1:4161"
	nsqdAddr         = "127.0.0.1:4150"
)

// FlattenType returns a map of flatten keys of the incoming dictionary, and
// values as the corresponding JSON types.
func FlattenType(d map[string]interface{}, p string) map[string]string {
	out := make(map[string]string)
	for key, value := range d {
		new_p := ""
		if len(p) > 0 {
			new_p = p + "." + key
		} else {
			new_p = key
		}
		if value == nil {
			// got JSON type null
			out[key] = "null"
		} else if reflect.TypeOf(value).Kind() == reflect.Map {
			// got an object
			s, ok := value.(map[string]interface{})
			if ok {
				for k, v := range FlattenType(s, new_p) {
					out[k] = v
				}
			} else {
				log.Fatalf("expected type map, got something else instead. key=%s, s=%s", key, s)
			}
		} else if reflect.TypeOf(value).Kind() == reflect.Slice {
			// got an array
			new_p += ".[]"
			s, ok := value.([]interface{})
			if ok {
				for _, d2 := range s {
					if reflect.TypeOf(d2).Kind() == reflect.Map {
						s2, ok := d2.(map[string]interface{})
						if ok {
							for k, v := range FlattenType(s2, new_p) {
								out[k] = v
							}
						} else {
							log.Fatalf("expected type map, got something else instead. key=%s, s2=%s", key, s2)
						}
					} else {
						// array here contains non-objects, so just save element type and break
						// note JSON doesn't require arrays have uniform type, but we'll assume it does
						out[key] = "Array[ " + PrettyPrintJsonType(d2) + " ]"
						break
					}
				}
			} else {
				log.Fatalf("expected type []interface{}, got something else instead. key=%s, s=%s", key, s)
			}
		} else {
			// got a basic type: Number, Boolean, or String
			out[new_p] = PrettyPrintJsonType(value)
		}
	}
	return out
}

// PrettyPrintJsonType accepts a variable (of type interface{}) and
// returns a human-readable string of "Number", "Boolean", "String", or "UNKNOWN".
func PrettyPrintJsonType(value interface{}) string {
	switch t := value.(type) {
	case float64:
		return "Number"
	case bool:
		return "Boolean"
	case string:
		return "String"
	default:
		log.Fatalf("unexpected type %T", t)
	}
	return "UNKNOWN"
}

// ConvertMapToJson simply takes a map of strings to strings,
// and converts it to a simplejson.Json object.
func ConvertMapToJson(m map[string]string) simplejson.Json {
	msg, _ := simplejson.NewJson([]byte("{}"))
	for k, v := range m {
		msg.Set(k, v)
	}
	return *msg
}

// InferType reads from an incoming channel msgChan, flattens and
// types the event, and puts it on another channel outChan.
func InferType(msgChan chan *nsq.Message, outChan chan simplejson.Json) {
	for {
		select {
		case m := <-msgChan:
			blob, err := simplejson.NewJson(m.Body)
			if err != nil {
				log.Fatalf(err.Error())
			}
			mblob, err := blob.Map()
			if err != nil {
				log.Fatalln(err)
			}
			flat := FlattenType(mblob, "")
			obj := ConvertMapToJson(flat)
			outChan <- obj
		}
	}
}

///// begin generic streamtools block code /////

type SyncHandler struct {
	msgChan chan *nsq.Message
}

func (self *SyncHandler) HandleMessage(m *nsq.Message) error {
	self.msgChan <- m
	return nil
}

func Writer(outChan chan simplejson.Json) {
	w := nsq.NewWriter(0)
	err := w.ConnectToNSQ(nsqdAddr)
	if err != nil {
		log.Fatal(err.Error())
	}
	for {
		select {
		case l := <-outChan:
			outMsg, _ := l.Encode()
			frameType, data, err := w.Publish(*outTopic, outMsg)
			if err != nil {
				log.Fatalf("frametype %d data %s error %s", frameType, string(data), err.Error())
			}
		}
	}

}

func main() {
	flag.Parse()
	channel := "type_inferencer"
	r, err := nsq.NewReader(*inTopic, channel)
	if err != nil {
		log.Println(*inTopic)
		log.Println(channel)
		log.Fatal(err.Error())
	}
	msgChan := make(chan *nsq.Message)
	outChan := make(chan simplejson.Json)
	go InferType(msgChan, outChan)
	go Writer(outChan)
	sh := SyncHandler{
		msgChan: msgChan,
	}
	r.AddHandler(&sh)
	_ = r.ConnectToLookupd(lookupdHTTPAddrs)
	<-r.ExitChan
}

///// end generic streamtools block code /////
