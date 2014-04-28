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
// credentials for authenticating with an SMTP server and the to, from and subject
// for the email message.
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
// By default, the block is configured for GMail.
func NewToEmail() blocks.BlockInterface {
	return &ToEmail{host: "smtp.gmail.com", port: 587, toPath: "to", fromPath: "from", subjectPath: "subject", msgPath: "msg"}
}

// Setup is called once before running the block. We build up the channels and specify what kind of block this is.
func (e *ToEmail) Setup() {
	e.Kind = "ToEmail"
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

// closeClient will attempt to quit or close the block's client.
func (e *ToEmail) closeClient() error {
	// quit, close and return
	var err error
	if err = e.client.Quit(); err != nil {
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

	e.username, err = util.ParseRequiredString(msgI, "Username")
	if err != nil {
		return err
	}

	e.password, err = util.ParseRequiredString(msgI, "Password")
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
func (e *ToEmail) Send(msg interface{}) error {
	// extract the 'to' and 'from' and build the email body
	from, to, email, err := e.buildEmail(msg)
	if err != nil {
		return err
	}

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

// Run is the block's main loop. Here we listen on the different channels we set up.
func (e *ToEmail) Run() {
	var err error
	for {
		err = nil
		select {
		case msgI := <-e.inrule:
			// get id/pw/host/port for SMTP
			if err = e.parseAuthRules(msgI); err != nil {
				e.Error(err.Error())
				continue
			}

			// get the to,from,subject for email
			if err = e.parseEmailRules(msgI); err != nil {
				e.Error(err.Error())
				continue
			}

			// if we already have a connection,  close it.
			if e.client != nil {
				if err = e.closeClient(); err != nil {
					e.Error(err.Error())
				} else {
					// give the connection a moment before reconnect
					time.Sleep(5 * time.Second)
				}
			}

			// initiate the SMTP connection and client
			if err = e.initClient(); err != nil {
				e.Error(err.Error())
				continue
			}

		case <-e.quit:
			// quit, close and return
			if err = e.closeClient(); err != nil {
				e.Error(err.Error())
			}
			return
		case msg := <-e.in:
			if e.client == nil {
				e.Error(fmt.Errorf("no smtp client available for toEmail block. please check the credentials."))
				continue
			}
			if err = e.Send(msg); err != nil {
				e.Error(err.Error())
				continue
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
	}
}
