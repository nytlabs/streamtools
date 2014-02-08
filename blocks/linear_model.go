package blocks

import (
	"github.com/nytlabs/gojee" // jee
	"log"
)

func LinearModel(b *Block) {

	type linearModelRule struct {
		Slope        float64
		Intercept    float64
		Variance     float64
		CovariateKey string
	}
	var rule *linearModelRule
	var tree *jee.TokenTree

	for {
		select {
		case m := <-b.Routes["set_rule"]:
			if rule == nil {
				rule = &linearModelRule{}
			}
			unmarshal(m, rule)
			// build the parser for the model
			token, err := jee.Lexer(rule.CovariateKey)
			if err != nil {
				log.Println(err.Error())
				break
			}
			tree, err = jee.Parser(token)
			if err != nil {
				log.Println(err.Error())
				break
			}
		case r := <-b.Routes["get_rule"]:
			if rule == nil {
				marshal(r, &linearModelRule{})
			} else {
				marshal(r, rule)
			}
		case msg := <-b.InChan:
			if rule == nil {
				break
			}
			if tree == nil {
				break
			}
			covariate, err := jee.Eval(tree, msg.Msg)
			if err != nil {
				log.Println(err.Error())
				break
			}
			// linear model
			X, ok := covariate.(float64)
			if !ok {
				log.Println("cannot type assert covariante to float64")
				break
			}
			Y := rule.Slope*X + rule.Intercept
			outMsg := map[string]interface{}{
				"ResponseVariable": Y,
			}
			out := BMsg{
				Msg: outMsg,
			}
			broadcast(b.OutChans, &out)
		case msg := <-b.AddChan:
			updateOutChans(msg, b)
		case <-b.QuitChan:
			quit(b)
			return
		}
	}
}
