package util

const (
    ERROR = iota
    WARN
    INFO
    DEBUG
    CREATE
    DELETE
    UPDATE
    QUERY
)

const (
    Reset = "\x1b[0m"
    Bright = "\x1b[1m"
    Dim = "\x1b[2m"
    Underscore = "\x1b[4m"
    Blink = "\x1b[5m"
    Reverse = "\x1b[7m"
    Hidden = "\x1b[8m"

    FgBlack = "\x1b[30m"
    FgRed = "\x1b[31m"
    FgGreen = "\x1b[32m"
    FgYellow = "\x1b[33m"
    FgBlue = "\x1b[34m"
    FgMagenta = "\x1b[35m"
    FgCyan = "\x1b[36m"
    FgWhite = "\x1b[37m"

    BgBlack = "\x1b[40m"
    BgRed = "\x1b[41m"
    BgGreen = "\x1b[42m"
    BgYellow = "\x1b[43m"
    BgBlue = "\x1b[44m"
    BgMagenta = "\x1b[45m"
    BgCyan = "\x1b[46m"
    BgWhite = "\x1b[47m"
)



const (
    VERSION = "0.2.1"
)

var LogInfo = map[int]string{
    0: FgRed + "ERROR" + Reset,
    1: FgYellow + "WARN" + Reset,
    2: FgWhite + "INFO" + Reset,
    3: BgMagenta + "DEBUG" + Reset,
    4: FgCyan + "CREATE" + Reset,
    5: FgCyan + "DELETE" + Reset,
    6: FgCyan + "UPDATE" + Reset,
    7: FgCyan + "QUERY " + Reset,
}

type LogMsg struct {
    Type int
    Data interface{}
    Id   string
}