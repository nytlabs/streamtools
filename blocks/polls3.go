package blocks

import (
	"bufio"
	"compress/gzip"
	"encoding/json"
	"launchpad.net/goamz/aws"
	"launchpad.net/goamz/s3"
	"log"
	"time"
)

func PollS3(b *Block) {

	type pollS3Rule struct {
		Period     float64
		BucketName string
		Prefix     string
		GzipFlag   bool
	}

	rule := &pollS3Rule{}

	unmarshal(<-b.Routes["set_rule"], &rule)

	log.Println("bucket", rule.BucketName)
	log.Println("prefix:", rule.Prefix)
	log.Println("gzip flag:", rule.GzipFlag)
	log.Println("poll interval:", rule.Period, "s")

	// The AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY environment variables are used.
	auth, err := aws.EnvAuth()
	if err != nil {
		panic(err.Error())
	}

	// triggerChan gets either the initial time or the ticker time, which is rule.Period
	// seconds after Now()
	triggerChan := make(chan time.Time, 1)

	samplePeriod := time.Duration(rule.Period) * time.Second
	ticker := time.NewTicker(samplePeriod)

	for {
		select {
		case t := <-ticker.C:
			// TODO this pattern introduces a bit of a delay. Is there a better way?
			triggerChan <- t
		case r := <-b.Routes["poll_now"]:
			triggerChan <- time.Now()
			log.Println("manual poll triggered")
			r.ResponseChan <- []byte(string(`{"PollS3 says":"Thanks Dude"}`))
		case t := <-triggerChan:
			log.Println("checking", rule.BucketName, ":", rule.Prefix)
			// Open Bucket
			s := s3.New(auth, aws.USEast)
			bucket := s.Bucket(rule.BucketName)
			// get the list
			list, err := bucket.List(rule.Prefix, "/", "", 2000)
			if err != nil {
				log.Fatal(err.Error())
			}
			log.Println("found", len(list.Contents), "files")
			for _, v := range list.Contents {
				lm, err := time.Parse("2006-01-02T15:04:05.000Z", v.LastModified)
				if err != nil {
					log.Fatal(err.Error())
				}
				if lm.After(t.Add(-samplePeriod)) {
					log.Println("[POLLS3] emitting", v.Key)
					br, _ := bucket.GetReader(v.Key)
					defer br.Close()

					var reader *bufio.Reader

					if rule.GzipFlag {
						gr, _ := gzip.NewReader(br)
						defer gr.Close()
						reader = bufio.NewReader(gr)
					} else {
						reader = bufio.NewReader(br)
					}
					for {
						line, err := reader.ReadBytes(10)
						if err != nil {
							log.Println(err.Error())
							break
						}
						var out BMsg
						json.Unmarshal(line, &out)
						broadcast(b.OutChans, out)
					}
				}
			}
			log.Println("done emitting")
		case r := <-b.Routes["set_rule"]:
			unmarshal(r, &rule)
			log.Println("got updated prefix:", rule.Prefix)
		case <-b.QuitChan:
			quit(b)
			return
		case msg := <-b.AddChan:
			updateOutChans(msg, b)
		}
	}
}
