package library

import (
	"fmt"
	"io"
	"net/smtp"
	"time"

	"github.com/nytlabs/streamtools/st/blocks"
	"github.com/nytlabs/streamtools/st/util"
)

// ToEmail holds channels we're going to use to communicate with streamtools,
// the SMTP credentials and client and the path for the to,
// from, subject and body of the email message to send.
type ToEmail struct {
	blocks.Block
	queryrule chan blocks.MsgChan
	inrule    blocks.MsgChan
	in        blocks.MsgChan
	quit      blocks.MsgChan

	host     string
	port     int
	username string
	password string

	toPath      string
	fromPath    string
	subjectPath string
	msgPath     string

	client *smtp.Client
}

// NewToEmail is a simple factory for streamtools to make new blocks of this kind.
// By default, the block is configured for Gmail.
func NewToEmail() blocks.BlockInterface {
	return &ToEmail{host: "smtp.gmail.com", port: 587, toPath: "to", fromPath: "from", subjectPath: "subject", msgPath: "msg"}
}

// Setup is called once before running the block. We build up the channels and specify what kind of block this is.
func (e *ToEmail) Setup() {
	e.Kind = "Network I/O"
	e.in = e.InRoute("in")
	e.inrule = e.InRoute("rule")
	e.queryrule = e.QueryRoute("rule")
	e.quit = e.Quit()
}

// initClient will create a new SMTP connection and set the block's client.
func (e *ToEmail) initClient() error {
	var err error
	e.client, err = newSMTPClient(e.username, e.password, e.host, e.port)
	if err != nil {
		return err
	}

	return nil
}

// reconnect will close the client, sleep the given duration and start a new connection.
func (e *ToEmail) reconnect(wait time.Duration) error {
	err := e.closeClient()
	if err != nil {
		e.Error(fmt.Sprintf("Problems closing SMTP client: %s", err.Error()))
	}
	// wait a moment before reconnecting
	time.Sleep(wait)

	return e.initClient()
}

// closeClient will attempt to quit or close the block's client.
func (e *ToEmail) closeClient() error {
	// quit, close and return
	var err error
	if e.client == nil {
		return nil
	}
	err = e.client.Quit()
	if err != nil {
		// quit failed. try a simple close
		err = e.client.Close()
	}
	return err
}

// newSMTPClient will connect, auth, say helo to the SMTP server and return the client.
func newSMTPClient(username, password, host string, port int) (*smtp.Client, error) {
	addr := fmt.Sprintf("%s:%d", host, port)
	client, err := smtp.Dial(addr)
	if err != nil {
		return client, err
	}

	// just saying HELO!
	if err = client.Hello("localhost"); err != nil {
		return client, err
	}

	// if the server can handle TLS, use it
	if ok, _ := client.Extension("STARTTLS"); ok {
		if err = client.StartTLS(nil); err != nil {
			return client, err
		}
	}

	// if the server can handle auth, use it
	if ok, _ := client.Extension("AUTH"); ok {
		auth := smtp.PlainAuth("", username, password, host)
		if err = client.Auth(auth); err != nil {
			return client, err
		}
	}

	return client, nil
}

// parseAuthInRules will expect a payload from the inrules channel and
// attempt to pull the SMTP auth credentials out it. If successful, this
// will also create and set the block's auth.
func (e *ToEmail) parseAuthRules(msgI interface{}) error {
	var err error
	e.host, err = util.ParseRequiredString(msgI, "Host")
	if err != nil {
		return err
	}

	e.port, err = util.ParseInt(msgI, "Port")
	if err != nil {
		return err
	}

	e.username, err = util.ParseString(msgI, "Username")
	if err != nil {
		return err
	}

	e.password, err = util.ParseString(msgI, "Password")
	if err != nil {
		return err
	}

	return nil
}

// parseEmailInRules will expect a payload from the inrules channel and
// attempt to pull and set the block's to, from and subject paths from it.
func (e *ToEmail) parseEmailRules(msgI interface{}) error {
	var err error
	e.toPath, err = util.ParseRequiredString(msgI, "ToPath")
	if err != nil {
		return err
	}

	e.fromPath, err = util.ParseRequiredString(msgI, "FromPath")
	if err != nil {
		return err
	}

	e.subjectPath, err = util.ParseString(msgI, "SubjectPath")
	if err != nil {
		return err
	}

	e.msgPath, err = util.ParseString(msgI, "MessagePath")
	if err != nil {
		return err
	}

	return nil
}

const emailTmpl = `From:%s
To:%s
Subject:%s

%s`

// buildEmail will attempt to pull the email's properties from the expected paths and
// put the email body together.
func (e *ToEmail) buildEmail(msg interface{}) (from, to string, email []byte, err error) {
	from, err = util.ParseString(msg, e.fromPath)
	if err != nil {
		return
	}
	to, err = util.ParseString(msg, e.toPath)
	if err != nil {
		return
	}
	var subject string
	subject, err = util.ParseString(msg, e.subjectPath)
	if err != nil {
		return
	}
	var body string
	body, err = util.ParseString(msg, e.msgPath)
	if err != nil {
		return
	}

	email = []byte(fmt.Sprintf(emailTmpl, from, to, subject, body))
	return
}

// Send will package and send the email.
func (e *ToEmail) send(from, to string, email []byte) error {
	var err error
	// set the 'from'
	if err = e.client.Mail(from); err != nil {
		return err
	}
	// set the 'to'
	if err = e.client.Rcpt(to); err != nil {
		return err
	}
	// get a handle of a writer for the message..
	var w io.WriteCloser
	if w, err = e.client.Data(); err != nil {
		return err
	}
	// ...and send the message body
	if _, err = w.Write(email); err != nil {
		return err
	}
	if err = w.Close(); err != nil {
		return err
	}

	return nil
}

// errWait is the increment time in seconds for waiting
// before reconnecting after encountering an error.
var errWait = 10

// normWait is the time duration to wait before performing a routine reconnect.
var normWait = 5 * time.Second

const maxRetries = 5

func (e *ToEmail) resetClient() bool {
	var err error
	wait := normWait
	connected := false
	// setup to retry reconnect if it fails
	for retries := 1; retries < maxRetries; retries++ {
		err = e.reconnect(wait)
		if err == nil {
			// if we succeeded, carry on.
			connected = true
			break
		}
		// incremental backoff if we failed first attempt
		wait = time.Duration(errWait*retries) * time.Second
		e.Error(fmt.Sprintf("Problems reconnecting to SMTP: %s. Trying again with a delay of %s", err.Error(), wait))
	}
	return connected
}

// Run is the block's main loop. Here we listen on the different channels we set up.
func (e *ToEmail) Run() {
	var err error
	var sent uint
	for {
		select {
		case msgI := <-e.inrule:
			// get id/pw/host/port for SMTP
			if err = e.parseAuthRules(msgI); err != nil {
				e.Error(fmt.Sprint("Unable to parse SMTP credentials: %s", err.Error()))
				continue
			}

			// get the to,from,subject for email
			if err = e.parseEmailRules(msgI); err != nil {
				e.Error(fmt.Sprintf("Unable to parse email component path rules: %s", err.Error()))
				continue
			}

			// if we don't have a client yet, initiate one.
			if e.client == nil {
				if err = e.initClient(); err != nil {
					e.Error(err)
				}
				continue
			}

			// if we do, start a new connection with new creds
			e.resetClient()
		case <-e.quit:
			if e.client != nil {
				if err = e.closeClient(); err != nil {
					e.Error(fmt.Sprintf("Unable to close SMTP connection: %s", err.Error()))
				}
			}
			return
		case msg := <-e.in:
			// if no client configured, error and give up.
			if e.client == nil {
				e.Error("The SMTP client does not exist yet. Please update the credentials.")
				continue
			}

			// extract the 'to' and 'from' and build the email body
			var email []byte
			var from, to string
			from, to, email, err = e.buildEmail(msg)
			if err != nil {
				e.Error(fmt.Sprintf("Unable to parse message for emailing: %s", err.Error()))
				continue
			}

			emlSent := false
			connected := true
			// give a few attempts to sending.
			for retries := 0; retries < maxRetries; retries++ {
				err = e.send(to, from, email)
				if err == nil {
					// we succeeded at sending the email. yay.
					emlSent = true
					sent++
					break
				}
				// attempt to reset client after each failure.
				connected = e.resetClient()
				if !connected {
					// if we cannot reconnect, dont retry sending.
					break
				}
				time.Sleep(time.Duration(retries*errWait) * time.Second)
			}
			if !emlSent {
				e.Error(fmt.Sprintf("Unable to send email: %s", err.Error()))
			}

			// reset the connection and the counter every 50 msgs or if theres been a send error.
			if (sent >= 50) || (err != nil) {
				sent = 0
				e.resetClient()
			}
		case MsgChan := <-e.queryrule:
			// deal with a query request
			MsgChan <- map[string]interface{}{
				"Host":     e.host,
				"Port":     e.port,
				"Username": e.username,
				"Password": e.password,

				"ToPath":      e.toPath,
				"FromPath":    e.fromPath,
				"SubjectPath": e.subjectPath,
				"MessagePath": e.msgPath,
			}
		}
		// reset
		err = nil
	}
}
