package library

import (
	"errors"
	"github.com/mjibson/go-dsp/fft"            // fft
	"github.com/nytlabs/gojee"                 // jee
	"github.com/nytlabs/streamtools/st/blocks" // blocks
	"github.com/nytlabs/streamtools/st/util"   // util
)

// specify those channels we're going to use to communicate with streamtools
type FFT struct {
	blocks.Block
	queryrule chan blocks.MsgChan
	queryfft  chan blocks.MsgChan
	inrule    blocks.MsgChan
	inpoll    blocks.MsgChan
	in        blocks.MsgChan
	out       blocks.MsgChan
	quit      blocks.MsgChan
}

func buildFFT(data tsData) [][]float64 {
	x := make([]float64, len(data.Values))
	for i, d := range data.Values {
		x[i] = d.Value
	}
	X := fft.FFTReal(x)
	Xout := make([][]float64, len(X))
	for i, Xi := range X {
		Xout[i] = make([]float64, 2)
		Xout[i][0] = real(Xi)
		Xout[i][1] = imag(Xi)
	}
	return Xout
}

// we need to build a simple factory so that streamtools can make new blocks of this kind
func NewFFT() blocks.BlockInterface {
	return &FFT{}
}

// Setup is called once before running the block. We build up the channels and specify what kind of block this is.
func (b *FFT) Setup() {
	b.Kind = "FFT"
	b.in = b.InRoute("in")
	b.inrule = b.InRoute("rule")
	b.queryrule = b.QueryRoute("rule")
	b.quit = b.Quit()
	b.out = b.Broadcast()
}

// Run is the block's main loop. Here we listen on the different channels we set up.
func (b *FFT) Run() {

	var err error
	//var path, lagStr string
	var path string
	var tree *jee.TokenTree
	//var lag time.Duration

	for {
		select {
		case ruleI := <-b.inrule:
			// set a parameter of the block
			rule, ok := ruleI.(map[string]interface{})
			if !ok {
				b.Error(errors.New("could not assert rule to map"))
			}
			path, err = util.ParseString(rule, "Path")
			if err != nil {
				b.Error(err)
				continue
			}
			tree, err = util.BuildTokenTree(path)
			if err != nil {
				b.Error(err)
				continue
			}
		case <-b.quit:
			// quit * time.Second the block
			return
		case msg := <-b.in:
			if tree == nil {
				continue
			}
			vI, err := jee.Eval(tree, msg)
			if err != nil {
				b.Error(err)
				continue
			}
			v, ok := vI.([]interface{})
			if !ok {
				b.Error(errors.New("could not assert timeseries to an array"))
				continue
			}
			values := make([]tsDataPoint, len(v))
			for i, vi := range v {
				value, ok := vi.(map[string]interface{})
				if !ok {
					b.Error(errors.New("could not assert value to map"))
					continue
				}
				tI, ok := value["timestamp"]
				if !ok {
					b.Error(errors.New("could not find timestamp in value"))
					continue
				}
				t, ok := tI.(float64)
				if !ok {
					b.Error(errors.New("could not assert timestamp to float"))
					continue
				}
				yI, ok := value["value"]
				if !ok {
					b.Error(errors.New("could not assert timeseries value to float"))
					continue
				}
				y, ok := yI.(float64)
				values[i] = tsDataPoint{
					Timestamp: t,
					Value:     y,
				}
			}
			data := tsData{
				Values: values,
			}
			out := map[string]interface{}{
				"fft": buildFFT(data),
			}
			b.out <- out
		case respChan := <-b.queryrule:
			// deal with a query request
			respChan <- map[string]interface{}{
				//"Window":     lagStr,
				"Path": path,
			}
		}
	}
}
