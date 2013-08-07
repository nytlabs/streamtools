package streamtools

// import (
// 	"github.com/bitly/go-simplejson"
// 	"github.com/bitly/nsq/nsq"
// 	"log"
// )

// func GetArrayLength(arrayKey string, msgChan chan *nsq.Message, outChan chan simplejson.Json) {
// 	for {
// 		select {
// 		case m := <-msgChan:
// 			blob, err := simplejson.NewJson(m.Body)
// 			if err != nil {
// 				log.Fatal(err.Error())
// 			}
// 			arr, err := blob.Get(arrayKey).Array()
// 			if err != nil {
// 				log.Fatal(err.Error())
// 			}
// 			l := len(arr)
// 			msg,_ := simplejson.NewJson([]byte("{}"))
//             msg.Set("len_" + arrayKey, l)
// 			outChan <- *msg
// 		}
// 	}
// }