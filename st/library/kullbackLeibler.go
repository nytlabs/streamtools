package library

import (
	"errors"
	"math"

	"github.com/nytlabs/gojee"                 // jee
	"github.com/nytlabs/streamtools/st/blocks" // blocks
	"github.com/nytlabs/streamtools/st/util"   // util
)

type KullbackLeibler struct {
	blocks.Block
	inrule    blocks.MsgChan
	queryrule chan blocks.MsgChan
	in        blocks.MsgChan
	out       blocks.MsgChan
	quit      blocks.MsgChan
}

func NewKullbackLeibler() blocks.BlockInterface {
	return &KullbackLeibler{}
}

func (b *KullbackLeibler) Setup() {
	b.Kind = "Stats"
	b.inrule = b.InRoute("rule")
	b.in = b.InRoute("in")
	b.queryrule = b.QueryRoute("rule")
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
	var qtree, ptree *jee.TokenTree
	var qpath, ppath string
	var err error
	for {
		select {
		case ruleI := <-b.inrule:
			qpath, err = util.ParseString(ruleI, "QPath")
			if err != nil {
				b.Error(err)
			}
			qtree, err = util.BuildTokenTree(qpath)
			ppath, err = util.ParseString(ruleI, "PPath")
			if err != nil {
				b.Error(err)
			}
			ptree, err = util.BuildTokenTree(ppath)
		case c := <-b.queryrule:
			c <- map[string]interface{}{
				"QPath": qpath,
				"PPath": ppath,
			}
		case <-b.quit:
			return
		case msg := <-b.in:
			if ptree == nil {
				continue
			}
			if qtree == nil {
				continue
			}
			pI, err := jee.Eval(ptree, msg)
			if err != nil {
				b.Error(err)
				break
			}
			qI, err := jee.Eval(qtree, msg)
			if err != nil {
				b.Error(err)
				break
			}
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
