package library

import (
	"encoding/csv"
	"io"
	"strconv"
	"strings"

	"github.com/nytlabs/gojee"
	// jee
	"github.com/nytlabs/streamtools/st/blocks" // blocks
	"github.com/nytlabs/streamtools/st/util"
)

// specify those channels we're going to use to communicate with streamtools
type ParseCSV struct {
	blocks.Block
	queryrule chan blocks.MsgChan
	inrule    blocks.MsgChan
	in        blocks.MsgChan
	inpoll    blocks.MsgChan
	out       blocks.MsgChan
	quit      blocks.MsgChan
}

// we need to build a simple factory so that streamtools can make new blocks of this kind
func NewParseCSV() blocks.BlockInterface {
	return &ParseCSV{}
}

// Setup is called once before running the block. We build up the channels and specify what kind of block this is.
func (b *ParseCSV) Setup() {
	b.Kind = "Parsers"
	b.Desc = "converts incoming CSV messages to JSON for use in streamtools"
	b.in = b.InRoute("in")
	b.inrule = b.InRoute("rule")
	b.inpoll = b.InRoute("poll")
	b.queryrule = b.QueryRoute("rule")
	b.quit = b.Quit()
	b.out = b.Broadcast()
}

// Run is the block's main loop. Here we listen on the different channels we set up.
func (b *ParseCSV) Run() {
	var tree *jee.TokenTree
	var path string
	var err error
	var headers []string
	var csvReader *csv.Reader

	for {
		select {
		case ruleI := <-b.inrule:
			// set a parameter of the block
			path, err = util.ParseString(ruleI, "Path")
			if err != nil {
				b.Error(err)
				continue
			}
			token, err := jee.Lexer(path)
			if err != nil {
				b.Error(err)
				continue
			}
			tree, err = jee.Parser(token)
			if err != nil {
				b.Error(err)
				continue
			}

			headers, err = util.ParseArrayString(ruleI, "Headers")
			if err != nil {
				b.Error(err)
				continue
			}
		case <-b.quit:
			// quit the block
			return
		case msg := <-b.in:
			// deal with inbound data
			if tree == nil {
				continue
			}
			var data string

			dataI, err := jee.Eval(tree, msg)
			if err != nil {
				b.Error(err)
				continue
			}

			switch value := dataI.(type) {
			case []byte:
				data = string(value[:])

			case string:
				data = value

			default:
				b.Error("data should be a string or a []byte")
				continue
			}

			csvReader = csv.NewReader(strings.NewReader(data))
			csvReader.TrimLeadingSpace = true
			// allow records to have variable numbers of fields
			csvReader.FieldsPerRecord = -1

		case <-b.inpoll:
			if csvReader == nil {
				b.Error("this block needs data to be pollable")
				break
			}
			record, err := csvReader.Read()
			if err != nil && err != io.EOF {
				b.Error(err)
				continue
			}
			row := make(map[string]interface{})
			for fieldIndex, field := range record {
				if fieldIndex >= len(headers) {
					row[strconv.Itoa(fieldIndex)] = field
				} else {
					header := headers[fieldIndex]
					row[header] = field
				}
			}

			b.out <- row

		case MsgChan := <-b.queryrule:
			// deal with a query request
			MsgChan <- map[string]interface{}{
				"Path":    path,
				"Headers": headers,
			}
		}
	}
}
