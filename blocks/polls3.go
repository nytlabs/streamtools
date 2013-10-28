package blocks

import (
	"bufio"
	"compress/gzip"
	"github.com/bitly/go-simplejson"
	"launchpad.net/goamz/aws"
	"launchpad.net/goamz/s3"
	"log"
	"time"
)

var (
	scanner *bufio.Scanner
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

	samplePeriod := time.Duration(rule.Period) * time.Second
	ticker := time.NewTicker(samplePeriod)

	for {
		select {
		case t := <-ticker.C:
			log.Println("checking", rule.BucketName, ":", rule.Prefix)
			// Open Bucket
			s := s3.New(auth, aws.USEast)
			bucket := s.Bucket(rule.BucketName)
			list, err := bucket.List(rule.Prefix, "/", "", 1000)
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
					if rule.GzipFlag {
						gr, _ := gzip.NewReader(br)
						defer gr.Close()
						scanner = bufio.NewScanner(gr)
					} else {
						scanner = bufio.NewScanner(br)
					}
					for scanner.Scan() {
						out, err := simplejson.NewJson(scanner.Bytes())
						if err != nil {
							log.Fatal(err.Error())
						}
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
