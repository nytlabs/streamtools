package library

import (
	"encoding/json"
	"fmt"
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

	to      string
	from    string
	subject string
}

// NewToEmail is a simple factory for streamtools to make new blocks of this kind.
// By default, the block is configured for GMail.
func NewToEmail() blocks.BlockInterface {
	return &ToEmail{host: "smtp.gmail.com", port: 587}
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
// attempt to pull and set the block's to, from and subject from it.
func (e *ToEmail) parseEmailRules(msgI interface{}) error {
	var err error
	e.to, err = util.ParseRequiredString(msgI, "To")
	if err != nil {
		return err
	}

	e.from, err = util.ParseRequiredString(msgI, "From")
	if err != nil {
		return err
	}

	e.subject, err = util.ParseString(msgI, "Subject")
	if err != nil {
		return err
	}

	return nil
}

const emailTmpl = `From:%s
To:%s
Subject:%s

%s`

func (e *ToEmail) buildEmail(msg interface{}) ([]byte, error) {
	// format the data for sending
	msgStr, err := json.Marshal(msg)
	if err != nil {
		return []byte{}, err
	}

	email := fmt.Sprintf(emailTmpl, e.from, e.to, e.subject, msgStr)
	return []byte(email), nil
}

func (e *ToEmail) Send(msg interface{}) error {
	// format the data for sending
	email, err := e.buildEmail(msg)
	if err != nil {
		return err
	}

	auth := smtp.PlainAuth("", e.username, e.password, e.host)
	return smtp.SendMail(fmt.Sprintf("%s:%d", e.host, e.port), auth, e.from, []string{e.to}, email)
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

				"To":      e.to,
				"From":    e.from,
				"Subject": e.subject,
			}
		}
	}
}
