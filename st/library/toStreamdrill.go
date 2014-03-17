package library

import (
	"github.com/nytlabs/gojee"
	"github.com/nytlabs/streamtools/st/blocks" // blocks
	"github.com/nytlabs/streamtools/st/util"
	"net/url"
	"net/http"
	"strings"
	"io"
	"fmt"
	"regexp"
)

// specify those channels we're going to use to communicate with streamtools
type ToStreamdrill struct {
	blocks.Block
	queryrule chan chan interface{}
	inrule    chan interface{}
	in        chan interface{}
	out       chan interface{}
	quit      chan interface{}
	url       string
	name      string
	entities  []string
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
	b.url = "http://localhost:9669"
}

// Run is the block's main loop. Here we listen on the different channels we set up.
func (b *ToStreamdrill) Run() {
	rd, wr := io.Pipe()
	client := &http.DefaultClient
	re := regexp.MustCompile("[\r\n]+")

	for {
		select {
		case msgI := <-b.inrule:
			urlS, _ := util.ParseString(msgI, "Url")
			name, _ := util.ParseString(msgI, "Name")
			entities, _ := util.ParseString(msgI, "Entities")


			if b.url != "" {
				b.Log("deleting existing trend")
				delReq, _ := http.NewRequest("DELETE", b.url + "/1/delete/" + b.name, nil)
				client.Do(delReq)
			} else {
				b.Log("deleting trend w/ same name trend (if it exists)")
				delReq, _ := http.NewRequest("DELETE", urlS + "/1/delete/" + name, nil)
				client.Do(delReq)
			}

			b.url = urlS
			b.name = name
			b.entities = strings.Split(entities, ":")
			b.Log(fmt.Sprintf("%d", len(b.entities)))

			createEntities := strings.Replace(strings.Join(b.entities, ":"), ".", "_", -1)
			callUrl := b.url + "/1/create/" + b.name + "/" + createEntities +"?size=1000000"
			_, err := client.Get(callUrl)
			if err != nil {
				b.Error(err)
			}
			b.Log("created new trend " + b.name)

			go func() {
				u, _ := url.Parse(b.url + "/1/update")

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

				b.Log("setting up streaming connection")

				r, streamErr := client.Do(req)
				if streamErr != nil {
					b.Error(streamErr)
				}
				defer r.Body.Close()
			}()

		case respChan := <-b.queryrule:
			// deal with a query request
			respChan <- map[string]interface{}{
				"Url":      b.url,
				"Name":      b.name,
				"Entities": strings.Join(b.entities, ":"),
			}
		case <-b.quit:
			wr.Close();
			// quit the block
			return
		case msg := <-b.in:
			var values []string

			for _, key := range b.entities {
				lexed, _ := jee.Lexer(key)
				parsed, _ := jee.Parser(lexed)
				e, err := jee.Eval(parsed, msg)
				if err != nil {
					b.Error(err)
					break
				}

				if e == nil || len(strings.TrimSpace(e.(string))) == 0 {
					break;
				}

				values = append(values, re.ReplaceAllString(e.(string), " "))
			}

			if len(values) == len(b.entities) {
				// put the into streamdrill
				message := b.name + "\t" + strings.Join(values, "\t") + "\n"
				wr.Write([]byte(message))
			}
		}
	}
}
