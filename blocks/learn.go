package blocks

import (
	"github.com/mikedewar/go-sgd" // sgd
	"github.com/nytlabs/gojee"    // jee
	"log"
)

func Learn(b *Block) {

	type learnRule struct {
		Features     []string
		Response     string
		Lossfunc     string
		Stepfunc     string
		InitialState []float64
	}

	rule := &learnRule{
		Lossfunc: "linear",
		Stepfunc: "inverse",
	}

	var featureTrees []*jee.TokenTree
	var responseTree *jee.TokenTree

	// sgd kernel
	dataChan := make(chan sgd.Obs)
	paramChan := make(chan sgd.Params)
	stateChan := make(chan chan []float64)
	kernelQuitChan := make(chan bool)
	lossfuncs := map[string]sgd.LossFunc{
		"linear":   sgd.GradLinearLoss,
		"logistic": sgd.GradLogisticLoss,
	}
	stepfuncs := map[string]sgd.StepFunc{
		"inverse": sgd.EtaInverse,
		"bottou":  sgd.EtaBottou,
	}
	kernelStarted := false

	for {
		select {
		case query := <-b.Routes["state"]:
			var model []float64
			if rule == nil {
				model = make([]float64, 0)
			} else {
				kernelRespChan := make(chan []float64)
				stateChan <- kernelRespChan
				model = <-kernelRespChan
			}
			marshal(query, model)
		case <-b.Routes["poll"]:
			var params []interface{}
			if (rule == nil) || (!kernelStarted) {
				params = make([]interface{}, 0)
			} else {
				kernelRespChan := make(chan []float64)
				stateChan <- kernelRespChan
				var model []float64
				model = <-kernelRespChan
				params = make([]interface{}, len(model))
				for i, p := range model {
					params[i] = p
				}

			}
			out := map[string]interface{}{
				"params": params,
			}
			outMsg := BMsg{
				Msg: out,
			}
			broadcast(b.OutChans, &outMsg)
		case ruleUpdate := <-b.Routes["set_rule"]:
			if rule == nil {
				rule = &learnRule{}
			}
			if kernelStarted {
				// if we already have a rule, then we've already started a
				// kernel, which we should now quit.
				kernelQuitChan <- true
			}
			unmarshal(ruleUpdate, rule)
			featureTrees = make([]*jee.TokenTree, len(rule.Features))

			for i, feature := range rule.Features {
				token, err := jee.Lexer(feature)
				if err != nil {
					log.Println(err.Error())
					break
				}
				tree, err := jee.Parser(token)
				if err != nil {
					log.Println(err.Error())
					break
				}
				featureTrees[i] = tree
			}
			if rule.Response == "" {
				break
			}
			token, err := jee.Lexer(rule.Response)
			if err != nil {
				log.Println(err.Error())
				break
			}
			responseTree, err = jee.Parser(token)
			if err != nil {
				log.Println(err.Error())
				break
			}
			grad := lossfuncs[rule.Lossfunc]
			step := stepfuncs[rule.Stepfunc]
			go sgd.SgdKernel(dataChan, paramChan, stateChan, kernelQuitChan, grad, step, rule.InitialState)
			kernelStarted = true

		case msg := <-b.Routes["get_rule"]:
			if rule == nil {
				marshal(msg, &learnRule{})
			} else {
				log.Println(rule)
				marshal(msg, rule)
			}

		case <-b.QuitChan:
			kernelQuitChan <- true
			quit(b)
			return

		case msg := <-b.AddChan:
			updateOutChans(msg, b)

		case msg := <-b.InChan:
			if rule == nil {
				break
			}
			if responseTree == nil {
				break
			}
			// get features
			x := make([]float64, len(featureTrees))
			for i, tree := range featureTrees {
				feature, err := jee.Eval(tree, msg.Msg)
				if err != nil {
					log.Println(err.Error())
					break
				}
				x[i] = feature.(float64)
			}
			// get response
			response, err := jee.Eval(responseTree, msg.Msg)
			if err != nil {
				log.Println(err.Error())
				break
			}
			y := response.(float64)
			// send to sgdkernel
			d := sgd.Obs{
				X: x,
				Y: y,
			}
			dataChan <- d
		}
	}
}
