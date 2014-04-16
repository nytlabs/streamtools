package library

import (
	"fmt"
	"log"
	"net/smtp"

	"github.com/nytlabs/streamtools/st/blocks"
	"github.com/nytlabs/streamtools/st/util"
)

// ToEmail holds channels we're going to use to communicate with streamtools,
// credentials for authenticating with an SMTP server and the to, from and subject
// for the email message.
type ToEmail struct {
	blocks.Block
	queryrule chan chan interface{}
	inrule    chan interface{}
	in        chan interface{}
	quit      chan interface{}

	host     string
	port     int
	username string
	password string

	toPath      string
	fromPath    string
	subjectPath string
	msgPath     string
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
		log.Printf("missing from: %s", e.fromPath)
		return
	}
	to, err = util.ParseString(msg, e.toPath)
	if err != nil {
		log.Printf("missing to: %s", e.toPath)
		return
	}
	var subject string
	subject, err = util.ParseString(msg, e.subjectPath)
	if err != nil {
		log.Printf("missing subject: %s", e.subjectPath)
		return
	}
	var body string
	body, err = util.ParseString(msg, e.msgPath)
	if err != nil {
		log.Printf("missing body: %s", e.msgPath)
		return
	}

	email = []byte(fmt.Sprintf(emailTmpl, from, to, subject, body))
	return
}

// Send will package and send the email.
func (e *ToEmail) Send(msg interface{}) error {
	// format the data for sending
	from, to, email, err := e.buildEmail(msg)
	if err != nil {
		return err
	}

	auth := smtp.PlainAuth("", e.username, e.password, e.host)
	return smtp.SendMail(fmt.Sprintf("%s:%d", e.host, e.port), auth, from, []string{to}, email)
}

// Run is the block's main loop. Here we listen on the different channels we set up.
func (e *ToEmail) Run() {
	var err error
	for {
		err = nil
		select {
		case msgI := <-e.inrule:
			// get id/pw/host/port for SMTP
			err = e.parseAuthRules(msgI)
			if err != nil {
				e.Error(err.Error())
				continue
			}

			// get the to,from,subject for email
			err = e.parseEmailRules(msgI)
			if err != nil {
				e.Error(err.Error())
				continue
			}
		case <-e.quit:
			return
		case msg := <-e.in:
			err = e.Send(msg)
			if err != nil {
				e.Error(err.Error())
				continue
			}
		case respChan := <-e.queryrule:
			// deal with a query request
			respChan <- map[string]interface{}{
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
