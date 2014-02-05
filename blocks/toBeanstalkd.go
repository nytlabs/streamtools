package blocks

import (
    /*
    "bufio"
    "errors"
    "fmt"
    "net"
    "os"
    "strings"
    */
    "log"
    "encoding/json"
    "github.com/nutrun/lentil"
)

/*
type Beanstalkd struct {
    conn   net.Conn
    reader *bufio.Reader
}

type Job struct {
    Id   uint64
    Body []byte
}
// Size of the reader buffer. Increase to handle large message bodies
var ReaderSize = 4096 // bufio.defaultSize
*/

func ToBeanstalkd(b *Block) {

    type toBeanstalkdRule struct {
        Host string
        Tube string
        TTR int
    }
    var conn lentil.Beanstalkd
    var tube = "default"
    var ttr = 0
    var rule *toBeanstalkdRule

    for {
        select {
        case m := <-b.Routes["set_rule"]:
            if rule == nil {
                rule = &toBeanstalkdRule{}
            }
            unmarshal(m, rule)
            conn, err := lentil.Dial(rule.Host)
            if err != nil {
                log.Panic(err.Error())
            } else {
                if len(rule.Tube) > 0 {
                    tube = rule.Tube
                }
                if rule.TTR > 0 {
                    ttr = rule.TTR
                }
                conn.Use(tube)
            }


        case r := <-b.Routes["get_rule"]:
            if rule == nil {
                marshal(r, &toBeanstalkdRule{})
            } else {
                marshal(r, rule)
            }

        case msg := <-b.InChan:
            if rule == nil {
                break
            }
            msgStr, err := json.Marshal(msg.Msg)
            if err != nil {
                log.Println("wow bad json")
            }
            /* your code goes here */
            jobId, err := conn.Put(0, 0, ttr, []byte(msgStr))
            if err != nil {
                log.Panic(err.Error())
            } else{
                log.Println("put jobId %d on queue", jobId)
            }
            //broadcast(b.OutChans, msg)
        case msg := <-b.AddChan:
            updateOutChans(msg, b)
        case <-b.QuitChan:
            quit(b)
            return
        }
    }
}


// Would love to use all these functions instead of lentil dependency
/*
func (this *Beanstalkd) send(format string, args ...interface{}) error {

    if Debug != nil {
        fmt.Fprintf(Debug, "(%v) -> ", this.conn)
        fmt.Fprintf(Debug, format, args...)
    }
    _, err := fmt.Fprintf(this.conn, format, args...)
    return err
}

func (this *Beanstalkd) recvline() (string, error) {
    reply, e := this.reader.ReadString('\n')
    if e != nil {
        return reply, e
    }
    if Debug != nil {
        fmt.Fprintf(Debug, "(%v) <- %v\n", this.conn, string(reply))
    }
    return reply, e
}

// Dial opens a connection to beanstalkd. The format of addr is 'host:port', e.g '0.0.0.0:11300'.
func Dial(addr string) (*Beanstalkd, error) {
    this := new(Beanstalkd)
    var e error
    this.conn, e = net.Dial("tcp", addr)
    if e != nil {
        return nil, e
    }
    this.reader = bufio.NewReaderSize(this.conn, ReaderSize)
    return this, nil
}

// Use is for producers. 
// Subsequent Put commands will put jobs into the tube specified by this command.
// If no use command has been issued, jobs will be put into the tube named "default".
func (this *Beanstalkd) Use(tube string) error {
    e := this.send("use %s\r\n", tube)
    if e != nil {
        return e
    }

    reply, e := this.recvline()
    if e != nil {
        return e
    }
    var usedTube string
    _, e = fmt.Sscanf(reply, "USING %s\r\n", &usedTube)
    if e != nil || tube != usedTube {
        return errors.New(reply)
    }
    return nil
}

// Put inserts a job into the queue.
func (this *Beanstalkd) Put(priority, delay, ttr int, data []byte) (uint64, error) {
    e := this.send("put %d %d %d %d\r\n%s\r\n", priority, delay, ttr, len(data), data)
    if e != nil {
        return 0, e
    }
    reply, e := this.recvline()
    if e != nil {
        return 0, e
    }
    var id uint64
    _, e = fmt.Sscanf(reply, "INSERTED %d\r\n", &id)
    if e != nil {
        return 0, errors.New(reply)
    }
    return id, nil
}

// Quit closes the connection to the queue.
func (this *Beanstalkd) Quit() error {
    e := this.send("quit\r\n")
    if e != nil {
        return e
    }
    return this.conn.Close()
}
*/