// Fakes a stream of data at your convenience.
// This contains a random time stamp, a random length array of random integers,
// and a nested json blob which in turn contains a random length utf8 string, an
// unsigned float, and a signed float.

package main

import (
	"bytes"
	"flag"
	"github.com/bitly/go-simplejson"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

var (
	topic       = flag.String("topic", "random", "nsq topic")
	jsonMsgPath = flag.String("file", "test.json", "json file to send")
	timeKey     = flag.String("key", "t", "key that holds time")

	// nsqHTTPAddrs = "127.0.0.1:4151"
	// raw_string   = "∰ ∈ ∉ ∌ ∂ According to ISO 10646-1:2000, sections D.7 and 2.3c, a device receiving UTF-8 shall interpret a \"malformed sequence in the same way that it interprets a character &amp;that is outside тіршілік-тынысының қолайлы дамыту, қауіпсіздігін, the adopted subset\" and “characters that are not within the adopted 남한산성(南漢山城)은 광주시, κόσμε 성남시, 산성으로, 784-16에 속해있다. subset shall be indicated to the user” by a receiving device. It can cost just $125 to buy a package of bees, and there&rsquo; is no real maintenance involved. Bees are typically bred in the south and shipped north in April, sent to beekeepers in a cage the size of a lunch box that can be mailed through the United States Postal Service. To buy a mature hive that is already producing honey, like the ones the Durst Organization has, can cost $300 to as high as $1,000. “People look at pogosto ne da razlikovati od nestrupenih vrst ter da žrtev na začetku nemalokrat zgleda neprizadeto, je treba v primeru ugriza takoj poiskati zdravniško pomoč moms and say, ‘oh they’re just low income and that’s it,'” she said. “I think parents have changed. विकिपीडिया  यह यथासम्भव निष्पक्ष दृष्टिकोण वाली सूOur income may still be low, but we’re more educated”"
)

func main() {
	flag.Parse()
	streamtools.ImportBlock( *inTopic, *outTopic, *channel, streamtools.RandomStream )
}

// func writer() {
// 	msgJson, _ := simplejson.NewJson([]byte("{}"))
// 	client := &http.Client{}

// 	c := time.Tick(5 * time.Second)
// 	r := rand.New(rand.NewSource(99))

// 	for now := range c {
// 		a := int64(r.Float64() * 10000000000)
// 		strTime := now.UnixNano() - a
// 		msgJson.Set(*timeKey, int64(strTime/1000000))

// 		msgJson.Set("a", 10)

// 		b := make([]int, rand.Intn(10))
// 		for i, _ := range b {
// 			b[i] = rand.Intn(100)
// 		}
// 		msgJson.Set("b", b)

// 		nestJson, _ := simplejson.NewJson([]byte("{}"))
// 		l := rand.Intn(20) + 10
// 		d := make([]string, l)
// 		string_bank := strings.Fields(raw_string)
// 		for i, _ := range d {
// 			d[i] = string_bank[rand.Intn(len(string_bank))]
// 		}
// 		nestJson.Set("d", strings.Join(d, " ")+".")
// 		nestJson.Set("e", rand.Float32()*8888)
// 		nestJson.Set("f", rand.Float32()-rand.Float32()*32)
// 		msgJson.Set("c", nestJson)

// 		outMsg, _ := msgJson.Encode()
// 		msgReader := bytes.NewReader(outMsg)
// 		resp, err := client.Post("http://"+nsqHTTPAddrs+"/put?topic="+*topic, "data/multi-part", msgReader)
// 		if err != nil {
// 			log.Fatalf(err.Error())
// 		}
// 		resp.Body.Close()
// 	}
// }

// func main() {

// 	flag.Parse()

// 	stopChan := make(chan int)

// 	go writer()

// 	<-stopChan
// }
