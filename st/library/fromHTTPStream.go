package library

import (
	"bytes"
	"encoding/json"
	"github.com/nytlabs/streamtools/st/blocks" // blocks
	"log"
	"net/http"
	"strings"
	"time"
)

// specify those channels we're going to use to communicate with streamtools
type FromHTTPStream struct {
	blocks.Block
	queryrule chan chan interface{}
	inrule    chan interface{}
	in        chan interface{}
	out       chan interface{}
	quit      chan interface{}
	endpoint  string
	auth      string
}

// a bit of boilerplate for streamtools
func NewFromHTTPStream() blocks.BlockInterface {
	return &FromHTTPStream{}
}

func (b *FromHTTPStream) Setup() {
	b.Kind = "FromHTTPStream"
	b.in = b.InRoute("in")
	b.inrule = b.InRoute("rule")
	b.queryrule = b.QueryRoute("rule")
	b.quit = b.InRoute("quit")
	b.out = b.Broadcast()
}

// creates a persistent HTTP connection, emitting all messages from
// the stream into streamtools
func (b *FromHTTPStream) Run() {

	var body bytes.Buffer
	// these are the possible delimiters
	d1 := []byte{125, 10} // this is }\n
	d2 := []byte{13, 10}  // this is CRLF

	client := &http.Client{}
	var res *http.Response

	for {
		select {
		case ruleI := <-b.inrule:
			rule := ruleI.(map[string]string)
			endpoint := rule["Endpoint"]
			auth := rule["Auth"]

			req, err := http.NewRequest("GET", endpoint, nil)

			if err != nil {
				log.Fatal(err)
			}
			if len(auth) > 0 {
				req.SetBasicAuth(strings.Split(auth, ":")[0], strings.Split(auth, ":")[1])
			}
			res, err = client.Do(req)
			if err != nil {
				log.Fatal(err)
			}
			defer res.Body.Close()

			b.endpoint = endpoint
			b.auth = auth

		case c := <-b.queryrule:
			c <- map[string]interface{}{
				"Endpoint": b.endpoint,
				"Auth":     b.auth,
			}
		case <-b.quit:
			// quit the block
			return

		default:
			if res == nil {
				time.Sleep(500 * time.Millisecond)
				break
			}
			buffer := make([]byte, 5*1024)
			p, err := res.Body.Read(buffer)

			if err != nil && err.Error() == "EOF" {
				log.Println("End of stream reached!")
				res = nil
				continue
			}

			if err != nil {
				log.Fatal(err.Error())
			}
			// catch odd little buffers
			if p < 2 {
				break
			}
			body.Write(buffer[:p])
			if bytes.Equal(d1, buffer[p-2:p]) { // ended with }\n
				for _, blob := range bytes.Split(body.Bytes(), []byte{10}) { // split on new line in case there are multuple messages per buffer
					if len(blob) > 0 {
						var outMsg interface{}
						err := json.Unmarshal(blob, &outMsg)
						if err != nil {
							log.Println("cannot unmarshal json")
							continue
						}
						b.out <- map[string]interface{}{
							"Msg": outMsg,
						}
					}
				}
				body.Reset()
			} else if bytes.Equal(d2, buffer[p-2:p]) { // ended with CRLF which we don't care about
				body.Reset()
			}
		}
	}
}
