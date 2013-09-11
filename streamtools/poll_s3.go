package streamtools

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

func PollS3(outChan chan *simplejson.Json, ruleChan chan *simplejson.Json) {

	rules := <-ruleChan
	d, err := rules.Get("duration").Int()
	if err != nil {
		log.Fatal(err.Error())
	}

	bucketName, err := rules.Get("bucketname").String()
	if err != nil {
		log.Fatal(err.Error())
	}

	prefix, err := rules.Get("prefix").String()
	if err != nil {
		log.Fatal(err.Error())
	}

	gzipFlag, err := rules.Get("gzipFlag").Bool()
	if err != nil {
		log.Fatal(err.Error())
	}

	log.Println("[POLLS3] bucket", bucketName)
	log.Println("[POLLS3] prefix:", prefix)
	log.Println("[POLLS3] gzip flag:", gzipFlag)
	log.Println("[POLLS3] poll interval:", d, "s")

	// The AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY environment variables are used.
	auth, err := aws.EnvAuth()
	if err != nil {
		panic(err.Error())
	}

	sampleDuration := time.Duration(d) * time.Second
	ticker := time.NewTicker(sampleDuration)

	for {
		select {
		case t := <-ticker.C:
			log.Println("[POLLS3] checking", bucketName, ":", prefix)
			// Open Bucket
			s := s3.New(auth, aws.USEast)
			bucket := s.Bucket(bucketName)
			list, err := bucket.List(prefix, "/", "", 1000)
			if err != nil {
				log.Fatal(err.Error())
			}
			log.Println("[POLLS3] found", len(list.Contents), "files")
			for _, v := range list.Contents {
				lm, err := time.Parse("2006-01-02T15:04:05.000Z", v.LastModified)
				if err != nil {
					log.Fatal(err.Error())
				}
				if lm.After(t.Add(-sampleDuration)) {
					log.Println("[POLLS3] emitting", v.Key)
					br, _ := bucket.GetReader(v.Key)
					defer br.Close()
					if gzipFlag {
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
						outChan <- out
					}
				}

			}
			log.Println("[POLLS3] done emitting")
		case rule := <-ruleChan:
			prefix, err = rule.Get("prefix").String()
			if err != nil {
				log.Fatal(err.Error())
			}
			log.Println("[POLLS3] got updated prefix:", prefix)
		}
	}

}
