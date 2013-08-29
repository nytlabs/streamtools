package main

import (
	"flag"
	"github.com/bitly/go-simplejson"
	"github.com/mikedewar/stream_tools/streamtools"
	"log"
)

var (
	pollInterval = flag.Float64("poll_interval", 300, "check S3 every X seconds")
	writeTopic   = flag.String("write_topic", "", "streamtools topic to write to")
	bucket       = flag.String("bucket", "", "s3 bucket")
	prefix       = flag.String("prefix", "", "s3 prefix")
	gzip         = flag.Bool("gzip_flag", true, "set to false if the files on S3 aren't gzipped")
)

func main() {
	log.SetFlags(log.Lshortfile)
	flag.Parse()
	S3PollBlock := streamtools.NewOutBlock(streamtools.PollS3, "S3Poller")
	rule, err := simplejson.NewJson([]byte("{}"))
	if err != nil {
		log.Fatal(err.Error())
	}
	rule.Set("duration", *pollInterval)
	rule.Set("bucketname", *bucket)
	rule.Set("prefix", *prefix)
	rule.Set("gzipFlag", *gzip)
	S3PollBlock.RuleChan <- *rule
	S3PollBlock.Run(*writeTopic, "8080")
}
