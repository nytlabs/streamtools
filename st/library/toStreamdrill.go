package library

import (
	"github.com/nytlabs/gojee"
	"github.com/nytlabs/streamtools/st/blocks" // blocks
	"net/url"
	"net/http"
	"strings"
	"io"
	"regexp"
	"fmt"
)

// specify those channels we're going to use to communicate with streamtools
type ToStreamdrill struct {
	blocks.Block
	queryrule chan chan interface{}
	inrule    chan interface{}
	in        chan interface{}
	out       chan interface{}
	quit      chan interface{}
}

// we need to build a simple factory so that streamtools can make new blocks of this kind
func NewToStreamdrill() blocks.BlockInterface {
	return &ToStreamdrill{}
}

// Setup is called once before running the block. We build up the channels and specify what kind of block this is.
func (b *ToStreamdrill) Setup() {
	b.Kind = "ToStreamdrill"
	b.in = b.InRoute("in")
	b.inrule = b.InRoute("rule")
	b.queryrule = b.QueryRoute("rule")
	b.quit = b.Quit()
	b.out = b.Broadcast()
}

// Run is the block's main loop. Here we listen on the different channels we set up.
func (b *ToStreamdrill) Run() {
	rd, wr := io.Pipe()
	client := &http.DefaultClient
	re := regexp.MustCompile("[\r\n]+")

	// we use this variable for tests
	var ok bool

	var baseUrl     string = "http://localhost:9669"
	var name        string
	var entities    string
	var value       string
	var timestamp   string

	// internal parsed data
	var splitEntities []string
	//var parsedEntities []*jee.TokenTree
	var parsedValue *jee.TokenTree
	var parsedTimestamp *jee.TokenTree

	for {
		select {
		case ruleI := <-b.inrule:
			rule := ruleI.(map[string]interface{})

			// setup base baseUrl
			baseUrl, ok = rule["Url"].(string)
			if !ok {
				b.Error("bad baseUrl")
				break
			}
			// check parsability
			_, err := url.Parse(baseUrl)
			if err != nil {
				b.Error("unparsable baseUrl: " + baseUrl)
				break;
			}

			// setup trend name
			name, ok = rule["Name"].(string)
			if !ok {
				b.Error("must provide a name")
				break
			} else {
				// ensure we start fresh after a change
				b.Log("deleting trend '" + name + "' w/ same name trend (if it exists)")
				delReq, _ := http.NewRequest("DELETE", baseUrl + "/1/delete/" + name, nil)
				res, _ := client.Do(delReq)
				res.Body.Close()
			}

			// setup entities
			entities, ok = rule["Entities"].(string)
			if !ok {
				b.Error("must provide entities")
				break
			} else {
				splitEntities = strings.Split(entities, ":")

				// create the trend on streamdrill
				createUrl := baseUrl + "/1/create/" + name + "/" + strings.Replace(entities, ".", "_", -1) + "?size=1000000"
				_, err := client.Get(createUrl)
				if err != nil {
					b.Error("creating the trend failed")
					break
				}
				b.Log("created new trend '" + name + "'")
			}

			// (optionall) setup a value parameter
			value, ok = rule["Value"].(string)
			if ok {
				lexed, lexErr := jee.Lexer(value)
				if lexErr != nil {
					b.Error(lexErr)
					continue
				}
				parsed, parseErr := jee.Parser(lexed)
				if parseErr != nil {
					b.Error(parseErr)
					continue
				}
				parsedValue = parsed
			}

			// in case the timestamp should be set explicitely
			timestamp, ok = rule["Timestamp"].(string)
			if ok {
				lexed, lexErr := jee.Lexer(timestamp)
				if lexErr != nil {
					b.Error(lexErr)
					continue
				}
				parsed, parseErr := jee.Parser(lexed)
				if parseErr != nil {
					b.Error(parseErr)
					continue
				}
				parsedTimestamp = parsed
			}

			go func() {
				u, err := url.Parse(baseUrl + "/1/update")
				if err == nil {
					req := &http.Request{
						Method:           "POST",
						ProtoMajor:       1,
						ProtoMinor:       1,
						URL:              u,
						TransferEncoding: []string{"chunked"},
						Header:           map[string][]string{},
						Body:             rd,
					}
					req.Header.Set("Content-type", "text/tab-separated-values")

					b.Log("setting up streaming connection to " + u.String() + " for trend '" + name + "'")

					r, streamErr := client.Do(req)
					if streamErr != nil {
						b.Error(streamErr)
					}
					defer r.Body.Close()
				} else {
					b.Error(err)
					b.Error("can't establish streaming connection")
				}
			}()

		case respChan := <-b.queryrule:
			// deal with a query request
			respChan <- map[string]interface{}{
				"Url":       baseUrl,
				"Name":      name,
				"Entities":  entities,
				"Value":     value,
				"Timestamp": timestamp,
			}
		case <-b.quit:
			wr.Close();
			// quit the block
			return
		case msg := <-b.in:
			var values []string

			for _, key := range splitEntities {
				lexed, _ := jee.Lexer(key)
				parsed, _ := jee.Parser(lexed)
				e, err := jee.Eval(parsed, msg)
				if err != nil {
					b.Error(err)
					continue
				}

				if e == nil || len(strings.TrimSpace(e.(string))) == 0 {
					continue
				}

				values = append(values, re.ReplaceAllString(e.(string), " "))
			}



			if len(values) == len(splitEntities) {
				// put the into streamdrill
				message := name + "\t" + strings.Join(values, "\t")
				if parsedValue != nil {
					v, err := jee.Eval(parsedValue, msg)
					if err == nil {
						message = fmt.Sprintf("%s\tv=%f", message, v)
					} else {
						b.Error(err)
					}
				}
				if parsedTimestamp != nil {
					v, err := jee.Eval(parsedTimestamp, msg)
					if err == nil {
						message = fmt.Sprintf("%s\tts=%d", message, v)
					} else {
						b.Error(err)
					}
				}

				wr.Write([]byte(message + "\n"))
			}
		}
	}
}
