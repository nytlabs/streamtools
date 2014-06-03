package library

import (
	"errors"

	"github.com/jasoncapehart/go-sgd"          //sgd
	"github.com/nytlabs/gojee"                 // jee
	"github.com/nytlabs/streamtools/st/blocks" // blocks
	"github.com/nytlabs/streamtools/st/util"
)

type Learn struct {
	blocks.Block
	queryrule chan blocks.MsgChan
	inrule    blocks.MsgChan
	inpoll    blocks.MsgChan
	in        blocks.MsgChan
	out       blocks.MsgChan
	quit      blocks.MsgChan
}

// a bit of boilerplate for streamtools
func NewLearn() blocks.BlockInterface {
	return &Learn{}
}

func (b *Learn) Setup() {
	b.Kind = "Stats"
	b.Desc = "applies stochastic gradient descent to learn the relationship between features and response"
	b.in = b.InRoute("in")
	b.inrule = b.InRoute("rule")
	b.inpoll = b.InRoute("poll")
	b.queryrule = b.QueryRoute("rule")
	b.quit = b.Quit()
	b.out = b.Broadcast()
}

func (b *Learn) Run() {

	dataChan := make(chan sgd.Obs)
	paramChan := make(chan sgd.Params)
	stateChan := make(chan chan []float64)
	kernelQuitChan := make(chan bool)

	lossfuncs := map[string]sgd.LossFunc{
		"linear":   sgd.GradLinearLoss,
		"logistic": sgd.GradLogisticLoss,
	}
	stepfuncs := map[string]sgd.StepFunc{
		"inverse":  sgd.EtaInverse,
		"constant": sgd.EtaConstant,
		"bottou":   sgd.EtaBottou,
	}
	kernelStarted := false

	var responsePath, lossfuncString, stepfuncString string
	var featurePaths []string
	var θ_0 []float64
	var featureTrees []*jee.TokenTree
	var responseTree *jee.TokenTree
	var err error

	for {
	Loop:
		select {
		case rule := <-b.inrule:
			if kernelStarted {
				// if we already have a rule, then we've already started a
				// kernel, which we should now quit.
				kernelQuitChan <- true
			}

			featurePaths, err = util.ParseArrayString(rule, "FeaturePaths")
			if err != nil {
				b.Error(err)
				continue
			}
			featureTrees = make([]*jee.TokenTree, len(featurePaths))
			for i, path := range featurePaths {
				token, err := jee.Lexer(path)
				if err != nil {
					b.Error(err)
					break
				}
				tree, err := jee.Parser(token)
				if err != nil {
					b.Error(err)
					break
				}
				featureTrees[i] = tree
			}
			responsePath, err = util.ParseString(rule, "ResponsePath")
			if err != nil {
				b.Error(err)
				break
			}
			token, err := jee.Lexer(responsePath)
			if err != nil {
				b.Error(err)
				break
			}
			responseTree, err = jee.Parser(token)
			if err != nil {
				b.Error(err)
				break
			}
			lossfuncString, err = util.ParseString(rule, "Lossfunc")
			if err != nil {
				b.Error(err)
				break
			}
			stepfuncString, err = util.ParseString(rule, "Stepfunc")
			if err != nil {
				b.Error(err)
				break
			}
			grad, ok := lossfuncs[lossfuncString]
			if !ok {
				b.Error(errors.New("Unknown loss function: " + lossfuncString))
			}
			step, ok := stepfuncs[stepfuncString]
			if !ok {
				b.Error(errors.New("Unknown step function: " + stepfuncString))
			}
			θ_0, err = util.ParseArrayFloat(rule, "InitialState")
			if err != nil {
				b.Error(err)
				break
			}
			go sgd.SgdKernel(dataChan, paramChan, stateChan, kernelQuitChan, grad, step, θ_0)
			kernelStarted = true

		case <-b.quit:
			kernelQuitChan <- true
			return
		case msg := <-b.in:
			if featureTrees == nil {
				continue
			}
			if responseTree == nil {
				continue
			}
			x := make([]float64, len(featureTrees))
			for i, tree := range featureTrees {
				feature, err := jee.Eval(tree, msg)
				if err != nil {
					b.Error(err)
					break Loop
				}
				fi, ok := feature.(float64)
				if !ok {
					b.Error(errors.New("features must be float64"))
					break Loop
				}
				x[i] = fi
			}
			responseI, err := jee.Eval(responseTree, msg)
			if err != nil {
				b.Error(err)
				break
			}
			y, ok := responseI.(float64)
			if !ok {
				b.Error(errors.New("response must be float64"))
				break
			}
			d := sgd.Obs{
				X: x,
				Y: y,
			}
			dataChan <- d
		case <-b.inpoll:
			var params []interface{}
			var model []float64
			if kernelStarted {
				kernelMsgChan := make(chan []float64)
				stateChan <- kernelMsgChan
				model = <-kernelMsgChan
				params = make([]interface{}, len(model))
				for i, p := range model {
					params[i] = p
				}
			}
			b.out <- map[string]interface{}{
				"params": params,
			}
		case c := <-b.queryrule:
			c <- map[string]interface{}{
				"Lossfunc":     lossfuncString,
				"Stepfunc":     stepfuncString,
				"FeaturePaths": featurePaths,
				"ResponsePath": responsePath,
				"InitialState": θ_0,
			}
		}
	}
}
