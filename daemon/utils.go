package daemon

import (
	"strconv"
    "encoding/json"
    "fmt"
    "net/http"
)

func IDService(idChan chan string) {
	i := 1
	for {
		id := strconv.Itoa(i)
		idChan <- id
		i += 1
	}
}

func ApiResponse(w http.ResponseWriter, statusCode int, statusTxt string) {
        response, err := json.Marshal(struct {
                StatusTxt  string      `json:"daemon"`
        }{
                statusTxt,
        })
        if err != nil {
                response = []byte(fmt.Sprintf(`{"daemon":"%s"}`, err.Error()))
        }

        w.Header().Set("Content-Type", "application/json; charset=utf-8")
        w.Header().Set("Content-Length", strconv.Itoa(len(response)))
        w.WriteHeader(statusCode)
        w.Write(response)
}

func DataResponse(w http.ResponseWriter, response []byte){
        w.Header().Set("Content-Type", "application/json; charset=utf-8")
        w.Header().Set("Content-Length", strconv.Itoa(len(response)))
        w.WriteHeader(200)
        w.Write(response)
}