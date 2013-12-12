package blocks

import (
	"log"
	"math/rand"
	"strings"
	"time"
)

var (
	raw_string = "∰ ∈ ∉ ∌ ∂ According to ISO 10646-1:2000, sections D.7 and 2.3c, a device receiving UTF-8 shall interpret a \"malformed sequence in the same way that it interprets a character &amp;that is outside тіршілік-тынысының қолайлы дамыту, қауіпсіздігін, the adopted subset\" and “characters that are not within the adopted 남한산성(南漢山城)은 광주시, κόσμε 성남시, 산성으로, 784-16에 속해있다. subset shall be indicated to the user” by a receiving device. It can cost just $125 to buy a package of bees, and there&rsquo; is no real maintenance involved. Bees are typically bred in the south and shipped north in April, sent to beekeepers in a cage the size of a lunch box that can be mailed through the United States Postal Service. To buy a mature hive that is already producing honey, like the ones the Durst Organization has, can cost $300 to as high as $1,000. “People look at pogosto ne da razlikovati od nestrupenih vrst ter da žrtev na začetku nemalokrat zgleda neprizadeto, je treba v primeru ugriza takoj poiskati zdravniško pomoč moms and say, ‘oh they’re just low income and that’s it,'” she said. “I think parents have changed. विकिपीडिया  यह यथासम्भव निष्पक्ष दृष्टिकोण वाली सूOur income may still be low, but we’re more educated”"
	options    = []string{"a", "b", "c", "d", "e", "f"}
)

func Random(b *Block) {

	type randomRule struct {
		Period int
	}

	rule := &randomRule{
		Period: 1,
	}

	c := time.Tick(time.Duration(rule.Period) * time.Second)
	r := rand.New(rand.NewSource(99))

	for {
		select {
		case r := <-b.Routes["set_rule"]:
			unmarshal(r, rule)
			log.Println("recieved new Period", rule.Period)
			c = time.Tick(time.Duration(rule.Period) * time.Second)
		case msg := <-b.Routes["get_rule"]:
			marshal(msg, rule)
		case now := <-c:
			msg := make(map[string]interface{})
			a := int64(r.Float64() * 10000000000)
			strTime := now.UnixNano() - a
			Set(msg, "t", int64(strTime/1000000))
			Set(msg, "a", 10)

			Set(msg, "random_int", rand.Intn(10)+1)

			randints := make([]int, rand.Intn(10))
			for i, _ := range randints {
				randints[i] = rand.Intn(100)
			}
			Set(msg, "random_integers", randints)

			random_float := r.Float64()
			Set(msg, "random_float", random_float)

			idx0 := rand.Intn(len(options))
			idx1 := rand.Intn(len(options))
			Set(msg, "option", options[idx0])

			if rand.Intn(10) > 5 {
				Set(msg, "sometimes.dot.option", options[rand.Intn(5)])
			}

			var nestJson interface{}
			nestJson = make(map[string]interface{})
			l := rand.Intn(20) + 10
			d := make([]string, l)
			string_bank := strings.Fields(raw_string)
			for i, _ := range d {
				d[i] = string_bank[rand.Intn(len(string_bank))]
			}
			Set(nestJson, "d", strings.Join(d, " ")+".")
			Set(nestJson, "e", rand.Float32()*8888)
			Set(nestJson, "f", rand.Float32()-rand.Float32()*32)
			Set(nestJson, "nestedOption", options[idx1])
			Set(msg, "c", nestJson)
			Set(msg, "e", rand.Float32()*8888)

			broadcast(b.OutChans, msg)
		case msg := <-b.AddChan:
			updateOutChans(msg, b)
		case <-b.QuitChan:
			quit(b)
			return
		}
	}
}
