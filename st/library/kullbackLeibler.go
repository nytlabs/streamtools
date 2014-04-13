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
	b.inA = b.InRoute("inA")
	b.inB = b.InRoute("inB")
	b.clear = b.InRoute("clear")
	b.quit = b.Quit()
	b.out = b.Broadcast()
}

func assertHistogram(hI interface{}) (map[string]float64, bool) {
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
	out := make(map[string]float64)
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
		out[k] = float64(v)
	}
	return out, ok
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
			p, ok := assertHistogram(pI)
			if !ok {
				b.Error(errors.New("p is not a Histogram"))
				continue
			}
			q, ok := assertHistogram(qI)
			if !ok {
				b.Error(errors.New("q is not a Histogram"))
				continue
			}
			outcomes := make([]string, 0, len(p)+len(q))
			for k := range p {
				outcomes = append(outcomes, k)
			}
			for k := range q {
				outcomes = append(outcomes, k)
			}
			pfull := make([]float64, len(outcomes))
			qfull := make([]float64, len(outcomes))
			for i, k := range outcomes {
				v, ok := p[k]
				if !ok {
					v = 0
				}
				pfull[i] = v
			}
			for i, k := range outcomes {
				v, ok := q[k]
				if !ok {
					v = 0
				}
				qfull[i] = v
			}
			kl := 0.0
			for i := range outcomes {
				kl += math.Log(pfull[i]/qfull[i]) * pfull[i]
			}
			b.out <- map[string]interface{}{
				"KL": kl,
			}
		}
	}
}
