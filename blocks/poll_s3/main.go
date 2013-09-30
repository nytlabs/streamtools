package main

import (
	"flag"
	"github.com/bitly/go-simplejson"
	"github.com/nytlabs/streamtools/streamtools"
	"log"
)

var (
	pollInterval = flag.Float64("poll-interval", 300, "check S3 every X seconds")
	writeTopic   = flag.String("write-topic", "", "streamtools topic to write to")
	bucket       = flag.String("bucket", "", "s3 bucket")
	prefix       = flag.String("prefix", "", "s3 prefix")
	gzip         = flag.Bool("gzip-flag", true, "set to false if the files on S3 aren't gzipped")
	rulePort     = flag.String("rule-port", "8080", "port to listen for new rules on")
	name         = flag.String("name", "poll-s3", "name of block")
)

func main() {
	flag.Parse()
	streamtools.SetupLogger(name)
	block := streamtools.NewOutBlock(streamtools.PollS3, *name)
	rule, err := simplejson.NewJson([]byte("{}"))
	if err != nil {
		log.Fatal(err.Error())
	}
	rule.Set("duration", *pollInterval)
	rule.Set("bucketname", *bucket)
	rule.Set("prefix", *prefix)
	rule.Set("gzipFlag", *gzip)
	block.RuleChan <- rule
	block.Run(*writeTopic, *rulePort)
}
