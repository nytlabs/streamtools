package macros

import "log"

var Macros = map[string]string{
	"annotation": "macro/annotation.html",
}

var MacroDefs = map[string][]byte{}

func Start() {
	for k, macro := range Macros {
		macroAsset, err := Asset(macro)
		if err != nil {
			log.Println("cannot find macro asset", macro)
			continue
		}
		MacroDefs[k] = macroAsset
	}
}
