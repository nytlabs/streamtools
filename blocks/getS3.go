package blocks

import (
	"bufio"
	"compress/gzip"
	"container/heap"
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
		retry  int
	}

	var reader *bufio.Reader
	var dumping bool

	rule := &getS3Rule{}

	// The AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY environment variables are used.
	auth, err := aws.EnvAuth()
	if err != nil {
		panic(err.Error())
	}

	// we make a priority queue to store keys in, in case Jacqui sends us a
	// billion keys to get all at once
	pq := &PriorityQueue{}
	heap.Init(pq)

	timer := time.NewTimer(time.Duration(10) * time.Second)
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
				timer.Reset(time.Duration(0))
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

			keyArray := getKeyValues(msg.Msg, "Key")
			if len(keyArray) == 0 {
				log.Println("No key found in message")
				break
			}
			keyInterface := keyArray[0]
			key := keyInterface.(string)

			bucketName := rule.BucketName
			j := &job{
				bucket: bucketName,
				key:    key,
				retry:  0,
			}
			queueMessage := &PQMessage{
				val: j,
				t:   time.Now(),
			}
			log.Println(j)
			heap.Push(pq, queueMessage)
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
		if len(*pq) == 0 {
			continue
		}
		// otherwise get on with the next key
		msgInterface := heap.Pop(pq)
		msg, ok := msgInterface.(*PQMessage)
		if !ok {
			log.Println("could not convert message to PQMessage")
			break
		}
		j, ok := msg.val.(*job)
		if !ok {
			log.Println("could not convert job interface to job object")
			break
		}
		// Open Bucket
		bucket := s.Bucket(j.bucket)
		log.Println("[POLLS3] emitting", j.bucket, j.key)
		br, err := bucket.GetReader(j.key)
		if err != nil {
			log.Println(err)
			j.retry++
			if j.retry < 3 {
				heap.Push(pq, msg)
			}
			continue // continue
		}
		defer br.Close()
		gr, err := gzip.NewReader(br)
		if err != nil {
			log.Println("failed to open a gzip reader")
			j.retry++
			if j.retry < 3 {
				heap.Push(pq, msg)
			}
			continue // continue
		}
		defer gr.Close()
		// set the reader
		reader = bufio.NewReader(gr)
		timer.Reset(time.Duration(0))
		dumping = true
	}
}
