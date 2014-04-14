package library

import (
	"errors"
	"github.com/nytlabs/streamtools/st/blocks" // blocks
	"math"
)

type KullbackLeibler struct {
	blocks.Block
	inA   chan interface{}
	inB   chan interface{}
	clear chan interface{}
	out   chan interface{}
	quit  chan interface{}
}

func NewKullbackLeibler() blocks.BlockInterface {
	return &KullbackLeibler{}
}

func (b *KullbackLeibler) Setup() {
	b.Kind = "KullbackLeibler"
	b.inA = b.InRoute("p")
	b.inB = b.InRoute("q")
	b.clear = b.InRoute("clear")
	b.quit = b.Quit()
	b.out = b.Broadcast()
}

var eps = 0.0001

type histogram map[string]float64

func newHistogram(hI interface{}) (histogram, bool) {
	h, ok := hI.(map[string]interface{})
	if !ok {
		return nil, ok
	}
	valuesI, ok := h["Histogram"]
	if !ok {
		return nil, ok
	}
	values, ok := valuesI.([]interface{})
	if !ok {
		return nil, ok
	}
	var out histogram
	out = make(map[string]float64)
	for _, valueI := range values {
		value, ok := valueI.(map[string]interface{})
		if !ok {
			return nil, ok
		}
		kI, ok := value["Label"]
		if !ok {
			return nil, ok
		}
		vI, ok := value["Count"]
		if !ok {
			return nil, ok
		}
		k, ok := kI.(string)
		if !ok {
			return nil, ok
		}
		v, ok := vI.(int)
		if !ok {
			return nil, ok
		}
		if v == 0 {
			out[k] = eps
		} else {
			out[k] = float64(v)
		}
	}
	z := 0.0
	for _, v := range out {
		z += float64(v)
	}
	for k, _ := range out {
		out[k] /= z
	}
	return out, ok
}

func (h histogram) normalise(p histogram) {
	for k, _ := range p {
		if _, ok := h[k]; !ok {
			h[k] = eps
		}
	}
	z := 0.0
	for _, v := range h {
		z += v
	}
	for k, v := range h {
		h[k] = v / z
	}
}

func (b *KullbackLeibler) Run() {
	A := make(chan interface{}, 1000)
	B := make(chan interface{}, 1000)
	for {
		select {
		case <-b.quit:
			return
		case msg := <-b.inA:
			select {
			case A <- msg:
			default:
				b.Error("the A queue is overflowing")
			}
		case msg := <-b.inB:
			select {
			case B <- msg:
			default:
				b.Error("the B queue is overflowing")
			}
		case <-b.clear:
		Clear:
			for {
				select {
				case <-A:
				case <-B:
				default:
					break Clear
				}
			}
		}
		for len(A) > 0 && len(B) > 0 {
			pI := <-A
			qI := <-B
			p, ok := newHistogram(pI)
			if !ok {
				b.Error(errors.New("p is not a Histogram"))
				continue
			}
			q, ok := newHistogram(qI)
			if !ok {
				b.Error(errors.New("q is not a Histogram"))
				continue
			}
			q.normalise(p)
			p.normalise(q)
			kl := 0.0
			for k, pi := range p {
				kl += math.Log(pi/q[k]) * pi
			}
			b.out <- map[string]interface{}{
				"KL": kl,
			}
		}
	}
}
