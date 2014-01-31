package blocks

import (
	"github.com/jacqui/gorecurses3/s3walker"
	"launchpad.net/goamz/aws"
	"log"
	"time"
)

// lists an S3 bucket, within a specified time inteval, starting with a
// specified prefix.
func ListS3(b *Block) {

	type listS3Rule struct {
		BucketName string
		Prefix     string
		Since      string
	}

	rule := &listS3Rule{}

	// The AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY environment variables are used.
	auth, err := aws.EnvAuth()
	if err != nil {
		panic(err.Error())
	}

	for {
		select {
		case <-b.InChan:
			out := make(map[string]interface{})
			// get the list
			listContents := s3walker.ListFiles(auth, rule.BucketName, rule.Prefix, "")

			log.Println("found", len(listContents), "files")
			outArray := []interface{}{}
			for _, v := range listContents {
				listelement := make(map[string]interface{})
				listelement["Key"] = v.Key
				if rule.Since == "" {
					outArray = append(outArray, listelement)
					continue
				}

				lm, err := time.Parse("2006-01-02T15:04:05.000Z", v.LastModified)
				if err != nil {
					log.Println(err.Error())
					break
				}
				since, err := time.ParseDuration(rule.Since)
				if err != nil {
					log.Println(err.Error())
					break
				}
				if lm.After(time.Now().Add(-since)) {
					listelement := make(map[string]interface{})
					listelement["Key"] = v.Key
					outArray = append(outArray, listelement)
				}
			}
			out["List"] = outArray
			outMsg := BMsg{
				Msg:          out,
				ResponseChan: nil,
			}
			broadcast(b.OutChans, &outMsg)
			log.Println("done emitting")
		case r := <-b.Routes["set_rule"]:
			unmarshal(r, rule)
			log.Println("got updated prefix:", rule.Prefix)
		case msg := <-b.Routes["get_rule"]:
			if rule == nil {
				marshal(msg, &listS3Rule{})
			} else {
				marshal(msg, rule)
			}
		case <-b.QuitChan:
			quit(b)
			return
		case msg := <-b.AddChan:
			updateOutChans(msg, b)
		}
	}
}
