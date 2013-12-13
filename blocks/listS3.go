package blocks

import (
	"launchpad.net/goamz/aws"
	"launchpad.net/goamz/s3"
	"log"
	"time"
)

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
			// Open Bucket
			s := s3.New(auth, aws.USEast)
			bucket := s.Bucket(rule.BucketName)
			// get the list
			list, err := bucket.List(rule.Prefix, "/", "", 2000)
			if err != nil {
				log.Println(err.Error())
			}
			log.Println("found", len(list.Contents), "files")
			outArray := []interface{}{}
			for _, v := range list.Contents {
				log.Println(v.Key)
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
					outArray = append(outArray, v)
				}
			}
			out["List"] = outArray
			broadcast(b.OutChans, out)
			log.Println("done emitting")
		case r := <-b.Routes["set_rule"]:
			unmarshal(r, &rule)
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
