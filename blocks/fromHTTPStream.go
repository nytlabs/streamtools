package blocks

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"
)

// creates a persistent HTTP connection, emitting all messages from
// the stream into streamtools
func FromHTTPStream(b *Block) {

	type longHTTPRule struct {
		Endpoint string
		Auth     string
	}
	var rule *longHTTPRule

	var body bytes.Buffer
	// these are the possible delimiters
	d1 := []byte{125, 10} // this is }\n
	d2 := []byte{13, 10}  // this is CRLF

	client := &http.Client{}
	var res *http.Response

	for {
		select {
		case m := <-b.Routes["set_rule"]:
			if rule == nil {
				rule = &longHTTPRule{}
			}
			unmarshal(m, rule)
			req, err := http.NewRequest("GET", rule.Endpoint, nil)
			if err != nil {
				log.Fatal(err)
			}
			if len(rule.Auth) > 0 {
				req.SetBasicAuth(strings.Split(rule.Auth, ":")[0], strings.Split(rule.Auth, ":")[1])
			}
			res, err = client.Do(req)
			if err != nil {
				log.Fatal(err)
			}
			defer res.Body.Close()

		case r := <-b.Routes["get_rule"]:
			if rule == nil {
				marshal(r, &longHTTPRule{})
			} else {
				marshal(r, rule)
			}
		case msg := <-b.AddChan:
			updateOutChans(msg, b)
		case <-b.QuitChan:
			quit(b)
			return
		default:
			if res == nil {
				time.Sleep(500 * time.Millisecond)
				break
			}
			buffer := make([]byte, 5*1024)
			p, err := res.Body.Read(buffer)
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
						out := BMsg{
							Msg: outMsg,
						}
						broadcast(b.OutChans, &out)
					}
				}
				body.Reset()
			} else if bytes.Equal(d2, buffer[p-2:p]) { // ended with CRLF which we don't care about
				body.Reset()
			}
		}
	}
}
