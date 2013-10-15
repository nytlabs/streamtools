package streamtools

import (
	"log"
	"strconv"
)

func IDService(idChan chan string) {
	i := 1
	for {
		id := strconv.Itoa(i)
		log.Println("generating new id:", id)
		idChan <- id
		i += 1
	}
}
