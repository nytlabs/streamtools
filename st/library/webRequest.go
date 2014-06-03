package library

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"

	"github.com/nytlabs/gojee"
	"github.com/nytlabs/streamtools/st/blocks" // blocks
	"github.com/nytlabs/streamtools/st/util"
)

func parseHeaders(headerRule map[string]interface{}) (map[string]string, error) {
	t := make(map[string]string)

	for k, v := range headerRule {
		switch r := v.(type) {
		case string:
			t[k] = r
		default:
			return nil, errors.New("value is not a string")
		}
	}

	return t, nil
}

// specify those channels we're going to use to communicate with streamtools
type WebRequest struct {
	blocks.Block
	queryrule chan blocks.MsgChan
	inrule    blocks.MsgChan
	inpoll    blocks.MsgChan
	in        blocks.MsgChan
	out       blocks.MsgChan
	quit      blocks.MsgChan
}

// we need to build a simple factory so that streamtools can make new blocks of this kind
func NewWebRequest() blocks.BlockInterface {
	return &WebRequest{}
}

// Setup is called once before running the block. We build up the channels and specify what kind of block this is.
func (b *WebRequest) Setup() {
	b.Kind = "Network I/O"
	b.Desc = "Makes requests to a given URL with specified HTTP method."
	b.in = b.InRoute("in")
	b.inrule = b.InRoute("rule")
	b.queryrule = b.QueryRoute("rule")
	b.out = b.Broadcast()
	b.quit = b.Quit()
}

// Run is the block's main loop. Here we listen on the different channels we set up.
func (b *WebRequest) Run() {
	var err error
	var ok bool

	url := ""
	var urlPath string
	var urlTree *jee.TokenTree

	bodyPath := "."
	var bodyTree *jee.TokenTree

	token, err := jee.Lexer(bodyPath)
	if err != nil {
		b.Error(err)
	}

	bodyTree, err = jee.Parser(token)
	if err != nil {
		b.Error(err)
	}

	var httpMethod string
	headerRule := map[string]interface{}{}
	headers, _ := parseHeaders(headerRule)

	transport := http.Transport{
		Dial: dialTimeout,
	}

	client := &http.Client{
		Transport: &transport,
	}

	for {
		select {
		case ruleI := <-b.inrule:
			httpMethod, err = util.ParseString(ruleI, "Method")
			if err != nil {
				b.Error(err)
				break
			}

			url, err = util.ParseString(ruleI, "Url")
			if err != nil {
				b.Error(err)
			}

			urlPath, err = util.ParseString(ruleI, "UrlPath")
			if err != nil {
				b.Error(err)
			}

			if len(url) != 0 && len(urlPath) != 0 {
				b.Error(errors.New("Specify either a url or a path to a url"))
				continue
			}

			if len(url) == 0 {
				token, err := jee.Lexer(urlPath)
				if err != nil {
					b.Error(err)
					continue
				}

				urlTree, err = jee.Parser(token)
				if err != nil {
					b.Error(err)
					continue
				}
			}

			bodyPath, err = util.ParseString(ruleI, "BodyPath")
			if err != nil {
				b.Error(err)
				continue
			}
			token, err := jee.Lexer(bodyPath)
			if err != nil {
				b.Error(err)
				continue
			}

			bodyTree, err = jee.Parser(token)
			if err != nil {
				b.Error(err)
				continue
			}

			rule := ruleI.(map[string]interface{})
			headerRuleI, ok := rule["Headers"]
			if !ok {
				continue
			}
			headerRule = headerRuleI.(map[string]interface{})
			p, err := parseHeaders(headerRule)
			if err == nil {
				headers = p
			} else {
				b.Error(err)
			}
		case <-b.quit:
			return

		case msg := <-b.in:
			var req *http.Request

			if urlTree != nil {
				urlInterface, err := jee.Eval(urlTree, msg)
				if err != nil {
					b.Error(err)
					continue
				}
				url, ok = urlInterface.(string)
				if !ok {
					b.Error(errors.New("couldn't assert url to a string"))
					continue
				}
			}

			if httpMethod == "POST" || httpMethod == "PUT" {
				bodyInterface, err := jee.Eval(bodyTree, msg)
				if err != nil {
					b.Error(err)
					continue
				}
				requestBody, err := json.Marshal(bodyInterface)
				if err != nil {
					b.Error(errors.New("couldn't marshal body"))
					continue
				}

				req, err = http.NewRequest(httpMethod, url, bytes.NewReader(requestBody))
				if err != nil {
					b.Error(err)
					break
				}

			} else {
				req, err = http.NewRequest(httpMethod, url, nil)
				if err != nil {
					b.Error(err)
					break
				}
			}

			for key, value := range headers {
				if key == "Host" {
					req.Host = value
				} else {
					req.Header.Set(key, value)
				}
			}

			resp, err := client.Do(req)
			if err != nil {
				b.Error(err)
				break
			}

			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				b.Error(err)
				break
			}

			var responseBody interface{}
			err = json.Unmarshal(body, &responseBody)
			if err != nil {
				responseBody = string(body)
			}

			// always return body/headers/status even if it's nasty xml
			outMsg := map[string]interface{}{
				"body":    responseBody,
				"headers": resp.Header,
				"status":  resp.Status,
			}

			resp.Body.Close()

			b.out <- outMsg

		case resp := <-b.queryrule:
			resp <- map[string]interface{}{
				"Url":      url,
				"UrlPath":  urlPath,
				"BodyPath": bodyPath,
				"Method":   httpMethod,
				"Headers":  headerRule,
			}
		}
	}
}
