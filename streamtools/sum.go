package streamtools

import (
	"bytes"
	"container/heap"
	"encoding/binary"
	"github.com/bitly/go-simplejson"
	"log"
	"time"
)

func Sum(inChan chan *simplejson.Json, ruleChan chan *simplejson.Json, queryChan chan stateQuery) {
	// block until we recieve a rule
	params := <-ruleChan
	windowSeconds, err := params.Get("window").Float64()
	if err != nil {
		log.Fatal(err.Error())
	}
	window := time.Duration(windowSeconds) * time.Second
	waitTimer := time.NewTimer(100 * time.Millisecond)

	key, err := params.Get("key").String()
	if err != nil {
		log.Fatal(err.Error())
	}

	pq := &PriorityQueue{}
	heap.Init(pq)

	// store the sum of the values in the PQ
	sum := 0.0

	for {
		select {
		case params = <-ruleChan:
		case query = <-queryChan:
			out, err := simplejson.NewJson([]byte("{}"))
			if err != nil {
				log.Fatal(err.Error())
			}
			out.Set("sum", sum)
			query.responseChan <- out
		case msg := <-inChan:
			val, err := getKey(key, msg).Float64()
			if err != nil {
				log.Fatal(err.Error())
			}
			log.Println("adding", val)
			sum += val

			// convert val to bytes for the queue
			buf := new(bytes.Buffer)
			err = binary.Write(buf, binary.LittleEndian, val)
			valBytes := buf.Bytes()

			queueMessage := &PQMessage{
				val: &valBytes,
				t:   time.Now(),
			}
			if err != nil {
				log.Fatal(err.Error())
			}
			heap.Push(pq, queueMessage)
		case <-waitTimer.C:
		}
		for {
			pqMsg, diff := pq.PeekAndShift(time.Now(), window)
			if pqMsg == nil {
				// either the queue is empty, or it's not time to emit
				if diff == 0 {
					diff = time.Duration(500) * time.Millisecond
				}
				waitTimer.Reset(diff)
				break
			}

			// decode the val so we can subtract it from the sum
			var val float64
			valBytes := pqMsg.(*PQMessage).val
			buf := bytes.NewBuffer(*valBytes)
			err := binary.Read(buf, binary.LittleEndian, &val)
			if err != nil {
				log.Println("binary.Read failed:", err)
			}
			log.Println("subtracting", val)
			sum -= val
		}
	}
}
