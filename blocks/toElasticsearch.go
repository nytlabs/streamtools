package blocks

import (
  "log"
  "github.com/mattbaird/elastigo/api"
  "github.com/mattbaird/elastigo/core"
)

// Posts a message to a specified Elasticsearch index with the given type.
func ToElasticsearch(b *Block) {

	type toElasticsearchRule struct {
		Host string
		Port   string
		Index  string
		Type   string
	}

	var rule *toElasticsearchRule

	// TODO check the endpoint for happiness
	for {
		select {
		case msg := <-b.Routes["set_rule"]:
			if rule == nil {
				rule = &toElasticsearchRule{}
			}
			unmarshal(msg, rule)
      // Set the Elasticsearch Host/Port to Connect to
      api.Domain = rule.Host
      api.Port = rule.Port

		case msg := <-b.Routes["get_rule"]:
			if rule == nil {
				marshal(msg, &toElasticsearchRule{})
			} else {
				marshal(msg, rule)
			}
		case <-b.QuitChan:
			quit(b)
			return
		case msg := <-b.InChan:
			if rule == nil {
				break
			}

      log.Println(msg.Msg)
      response, err := core.Index(true, rule.Index, rule.Type, "", msg.Msg)
			if err != nil {
				log.Println(err.Error())
				break
			}
      log.Printf("Index OK: %v", response.Ok)
  	}
	}
}
