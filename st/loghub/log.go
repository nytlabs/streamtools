package loghub

import (
	"encoding/json"
	"fmt"
	"log"
	"time"
)

const (
	ERROR = iota
	WARN
	INFO
	DEBUG
	CREATE
	DELETE
	UPDATE
	QUERY
	RULE_UPDATED
	UPDATE_RULE
	UPDATE_POSITION
	UPDATE_RATE
)

const (
	Reset      = "\x1b[0m"
	Bright     = "\x1b[1m"
	Dim        = "\x1b[2m"
	Underscore = "\x1b[4m"
	Blink      = "\x1b[5m"
	Reverse    = "\x1b[7m"
	Hidden     = "\x1b[8m"

	FgBlack   = "\x1b[30m"
	FgRed     = "\x1b[31m"
	FgGreen   = "\x1b[32m"
	FgYellow  = "\x1b[33m"
	FgBlue    = "\x1b[34m"
	FgMagenta = "\x1b[35m"
	FgCyan    = "\x1b[36m"
	FgWhite   = "\x1b[37m"

	BgBlack   = "\x1b[40m"
	BgRed     = "\x1b[41m"
	BgGreen   = "\x1b[42m"
	BgYellow  = "\x1b[43m"
	BgBlue    = "\x1b[44m"
	BgMagenta = "\x1b[45m"
	BgCyan    = "\x1b[46m"
	BgWhite   = "\x1b[47m"
)

var LogInfo = map[int]string{
	0:  "ERROR",
	1:  "WARN",
	2:  "INFO",
	3:  "DEBUG",
	4:  "CREATE",
	5:  "DELETE",
	6:  "UPDATE",
	7:  "QUERY ",
	8:  "RULE_UPDATED",
	9:  "UPDATE_RULE",
	10: "UPDATE_POSITION",
	11: "UPDATE_RATE",
}

var LogInfoColor = map[int]string{
	0: FgRed + "ERROR" + Reset,
	1: FgYellow + "WARN" + Reset,
	2: FgWhite + "INFO" + Reset,
	3: BgMagenta + "DEBUG" + Reset,
	4: FgCyan + "CREATE" + Reset,
	5: FgCyan + "DELETE" + Reset,
	6: FgCyan + "UPDATE" + Reset,
	7: FgCyan + "QUERY" + Reset,
	8: FgCyan + "UPDATE" + Reset,
}

type LogMsg struct {
	Type int
	Data interface{}
	Id   string
}

var Log chan *LogMsg
var UI chan *LogMsg
var AddLog chan chan []byte
var AddUI chan chan []byte

func Start() {
	Log = make(chan *LogMsg, 10)
	UI = make(chan *LogMsg, 10)
	AddLog = make(chan chan []byte)
	AddUI = make(chan chan []byte)
	go BroadcastStream()
}

// BroadcastStream routes logs and block system changes to websocket hubs
// and terminal.
func BroadcastStream() {
	var batch []interface{}

	var logOut []chan []byte
	var uiOut []chan []byte

	// we batch the logs every 50 ms so we can cut down on the amount
	// of messages we send
	dump := time.NewTicker(50 * time.Millisecond)

	for {
		select {
		case newUI := <-AddUI:
			uiOut = append(uiOut, newUI)
		case newLog := <-AddLog:
			logOut = append(logOut, newLog)
		case <-dump.C:
			if len(batch) == 0 {
				break
			}

			outBatch := struct {
				Log []interface{}
			}{
				batch,
			}

			joutBatch, err := json.Marshal(outBatch)
			if err != nil {
				log.Println("could not broadcast")
			}

			for _, v := range logOut {
				v <- joutBatch
			}

			batch = nil
		case l := <-Log:
			if l.Type == 0 {
				e, ok := l.Data.(error)
				if ok {
					l.Data = interface{}(e.Error())
				}
			}
			bclog := struct {
				Type string
				Data interface{}
				Id   string
			}{
				LogInfo[l.Type],
				l.Data,
				l.Id,
			}

			jsonData, err := json.Marshal(l.Data)
			if err != nil {
				log.Println("failed marshaling data into json")
				fmt.Println(fmt.Sprintf("%s [ %s ][ %s ] %s", time.Now().Format(time.Stamp), l.Id, LogInfoColor[l.Type], l.Data))
			} else {
				fmt.Println(fmt.Sprintf("%s [ %s ][ %s ] %s", time.Now().Format(time.Stamp), l.Id, LogInfoColor[l.Type], jsonData))
			}
			batch = append(batch, bclog)
		case l := <-UI:
			bclog := struct {
				Type string
				Data interface{}
				Id   string
			}{
				LogInfo[l.Type],
				l.Data,
				l.Id,
			}

			j, err := json.Marshal(bclog)
			if err != nil {
				log.Println("could not broadcast")
				break
			}

			for _, v := range uiOut {
				v <- j
			}
		}
	}
}
