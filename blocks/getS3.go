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

// Gets the key specified in the inbound message. Specify the bucket using a
// rule.
func GetS3(b *Block) {

	type getS3Rule struct {
		BucketName string
	}

	type job struct {
		bucket string
		key    string
	}

	var reader *bufio.Reader
	var dumping bool

	rule := &getS3Rule{}

	// The AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY environment variables are used.
	auth, err := aws.EnvAuth()
	if err != nil {
		panic(err.Error())
	}

	timer := time.NewTimer(time.Duration(10) * time.Second)
	todo := make(chan job, 100) // max number of jobs is 100
	s := s3.New(auth, aws.USEast)

	for {
		select {
		// timer will have a waiting message until we hit EOF. Each time this
		// case is called, one new message is emitted
		case <-timer.C:
			// this check should only be needed at the initialisation of the
			// block
			if !dumping {
				break
			}
			line, err := reader.ReadBytes(10)
			if err != nil {
				log.Println(err.Error())
				dumping = false
				break
			}
			var out interface{}
			err = json.Unmarshal(line, &out)
			if err != nil {
				log.Println(err)
				break
			}
			outMsg := BMsg{
				Msg:          out,
				ResponseChan: nil,
			}
			broadcast(b.OutChans, outMsg)
			timer.Reset(time.Duration(0))
			// the inChan case is responsible for putting a job into the bufferred
			// todo channel
		case msg := <-b.InChan:
			if rule == nil {
				log.Println("no rule set")
				break
			}
			/*
				bucketName := getKeyValues(msg, "bucketName")
				if len(bucketName) == 0 {
					log.Println("No bucket name found in message")
					break
				}
			*/
			keyArray := getKeyValues(msg.Msg, "Key")
			if len(keyArray) == 0 {
				log.Println("No key found in message")
				break
			}
			keyInterface := keyArray[0]
			key := keyInterface.(string)
			//bucket: bucketName[0].(string),
			bucketName := rule.BucketName
			j := job{
				bucket: bucketName,
				key:    key,
			}
			log.Println(j)
			//TODO this should be a priority queue
			if len(todo) == 100 {
				log.Println("Trying to queue up more than 100 keys!")
				break
			}
			todo <- j
		case r := <-b.Routes["set_rule"]:
			unmarshal(r, rule)
		case msg := <-b.Routes["get_rule"]:
			if rule == nil {
				marshal(msg, &getS3Rule{})
			} else {
				marshal(msg, &rule)
			}
		case <-b.QuitChan:
			quit(b)
			return
		case msg := <-b.AddChan:
			updateOutChans(msg, b)
		}
		// at this point either we've just emitted a message, finished a file,
		// loaded up a job, added an outChan or updated a rule. So we need to
		// check out what's going on.

		// if we're dumping then keep dumping
		if dumping {
			continue
		}
		// if there's nothing todo, then go wait for something to do
		if len(todo) == 0 {
			continue
		}
		// otherwise get on with the next key
		j := <-todo
		// Open Bucket
		bucket := s.Bucket(j.bucket)
		log.Println("[POLLS3] emitting", j.bucket, j.key)
		br, err := bucket.GetReader(j.key)
		if err != nil {
			log.Println(err)
			break
		}
		defer br.Close()
		gr, err := gzip.NewReader(br)
		if err != nil {
			log.Println("failed to open a gzip reader")
			break
		}
		defer gr.Close()
		// set the reader
		reader = bufio.NewReader(gr)
		timer.Reset(time.Duration(0))
		dumping = true
	}
}
