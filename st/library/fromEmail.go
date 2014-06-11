package library

import (
	"bytes"
	"crypto/tls"
	"net/mail"
	"time"

	"github.com/mxk/go-imap/imap"

	"github.com/nytlabs/streamtools/st/blocks"
	"github.com/nytlabs/streamtools/st/util"
)

// FromEmail holds channels we're going to use to communicate with streamtools,
// credentials for authenticating with an IMAP server and the IMAP client.
type FromEmail struct {
	blocks.Block
	queryrule chan blocks.MsgChan
	inrule    blocks.MsgChan
	out       blocks.MsgChan
	quit      blocks.MsgChan

	host     string
	username string
	password string
	mailbox  string

	client *imap.Client
	idling bool
}

// NewFromEmail is a simple factory for streamtools to make new blocks of this kind.
// By default, the block is configured for GMail.
func NewFromEmail() blocks.BlockInterface {
	return &FromEmail{host: "imap.gmail.com", mailbox: "INBOX"}
}

// newIMAPClient will initiate a new IMAP connection with the given creds.
func newIMAPClient(host, username, password, mailbox string) (*imap.Client, error) {
	client, err := imap.DialTLS(host, new(tls.Config))
	if err != nil {
		return client, err
	}

	_, err = client.Login(username, password)
	if err != nil {
		return client, err
	}

	_, err = imap.Wait(client.Select(mailbox, false))
	if err != nil {
		return client, err
	}

	return client, nil
}

// idle will initiate an IMAP idle and wait for updates. Any time the connection finds a idle update,
// it will terminate the idle, fetch any unread email messages and kick idle off again. Every 20
// minutes, it will reset the idle to keep it alive.
func (e *FromEmail) idle() {
	// keep track of an ongoing idle
	e.idling = true
	defer func() { e.idling = false }()

	var err error
	_, err = e.client.Idle()
	if err != nil {
		e.Error(err.Error())
		return
	}

	// kicks off occasional data check during Idle
	poll := make(chan uint, 1)
	poll <- 0

	// setup ticker to reset the idle every 20 minutes (RFC-2177 recommends every <=29 mins)
	reset := time.NewTicker(20 * time.Minute)

	for {
		select {
		case <-poll:
			// attempt to fill pipe with new data
			err = e.client.Recv(0)
			if err != nil {
				// imap.ErrTimeout here means 'no data available'
				if err == imap.ErrTimeout {
					sleep(poll)
					continue
				} else {
					e.Error(err.Error())
					return
				}
			}

			// check the pipe for data
			if len(e.client.Data) > 0 {
				// term idle and fetch unread
				_, err = e.client.IdleTerm()
				if err != nil {
					e.Error(err.Error())
					sleep(poll)
					return
				}
				e.idling = false

				// put any new unread messages on the channel
				err = e.fetchUnread()
				if err != nil {
					e.Error(err.Error())
					sleep(poll)
					return
				}

				// kick off that idle again
				_, err = e.client.Idle()
				if err != nil {
					e.Error(err.Error())
					sleep(poll)
					return
				}
				e.idling = true
			}
			// clean the pipe
			e.client.Data = nil
			// sleep a bit before checking the pipe again
			sleep(poll)

		case <-reset.C:
			_, err = e.client.IdleTerm()
			if err != nil {
				e.Error(err.Error())
				return
			}

			_, err = e.client.Idle()
			if err != nil {
				e.Error(err.Error())
				return
			}
		}
	}
}

func sleep(poll chan uint) {
	go func() {
		time.Sleep(10 * time.Second)
		poll <- 1
	}()
}

// fetchUnread emails will check the current mailbox for any unread messages. If it finds
// some, it will grab the email bodies, parse them and pass them along the block's out channel.
func (e *FromEmail) fetchUnread() error {
	cmd, err := findUnreadEmails(e.client)
	if err != nil {
		return err
	}

	var emails []map[string]interface{}
	emails, err = getEmails(e.client, cmd)
	if err != nil {
		return err
	}

	for _, email := range emails {
		e.out <- email
	}

	return nil
}

// getEmails will fetch the full bodies of all emails listed in the given command.
func getEmails(client *imap.Client, cmd *imap.Command) ([]map[string]interface{}, error) {
	var emails []map[string]interface{}
	seq := new(imap.SeqSet)
	for _, rsp := range cmd.Data {
		for _, uid := range rsp.SearchResults() {
			seq.AddNum(uid)
		}
	}
	if seq.Empty() {
		return emails, nil
	}
	fCmd, err := imap.Wait(client.UIDFetch(seq, "INTERNALDATE", "BODY[]", "UID", "RFC822.HEADER"))
	if err != nil {
		return emails, err
	}

	var email map[string]interface{}
	for _, msgData := range fCmd.Data {
		msgFields := msgData.MessageInfo().Attrs
		email, err = newEmailMessage(msgFields)
		if err != nil {
			return emails, err
		}
		emails = append(emails, email)

		// mark message as read
		fSeq := new(imap.SeqSet)
		fSeq.AddNum(imap.AsNumber(msgFields["UID"]))
		_, err = imap.Wait(client.UIDStore(fSeq, "+FLAGS", "\\SEEN"))
		if err != nil {
			return emails, err
		}
	}
	return emails, nil
}

// newEmailMessage will parse an imap.FieldMap into an map[string]interface{}. This
// will expect the message to container the internaldate and the body with
// all headers included.
func newEmailMessage(msgFields imap.FieldMap) (map[string]interface{}, error) {
	var email map[string]interface{}
	// parse the header
	rawHeader := imap.AsBytes(msgFields["RFC822.HEADER"])
	msg, err := mail.ReadMessage(bytes.NewReader(rawHeader))
	if err != nil {
		return email, err
	}

	email = map[string]interface{}{
		"internal_date": imap.AsDateTime(msgFields["INTERNALDATE"]),
		"body":          imap.AsString(msgFields["BODY[]"]),
		"from":          msg.Header.Get("From"),
		"to":            msg.Header.Get("To"),
		"subject":       msg.Header.Get("Subject"),
	}

	return email, nil
}

// findUnreadEmails will run a find the UIDs of any unread emails in the
// mailbox.
func findUnreadEmails(conn *imap.Client) (*imap.Command, error) {
	// get headers and UID for UnSeen message in src inbox...
	cmd, err := imap.Wait(conn.UIDSearch("UNSEEN"))
	if err != nil {
		return &imap.Command{}, err
	}
	return cmd, nil
}

// Setup is called once before running the block. We build up the channels and specify what kind of block this is.
func (e *FromEmail) Setup() {
	e.Kind = "Network I/O"
	e.out = e.Broadcast()
	e.inrule = e.InRoute("rule")
	e.queryrule = e.QueryRoute("rule")
	e.quit = e.Quit()
}

// parseAuthInRules will expect a payload from the inrules channel and
// attempt to pull the IMAP auth credentials out it.
func (e *FromEmail) parseAuthRules(msgI interface{}) error {
	var err error
	e.host, err = util.ParseRequiredString(msgI, "Host")
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

	e.mailbox, err = util.ParseRequiredString(msgI, "Mailbox")
	if err != nil {
		return err
	}

	return nil
}

func (e *FromEmail) initClient() error {
	// initiate IMAP client with new creds
	var err error
	e.client, err = newIMAPClient(e.host, e.username, e.password, e.mailbox)
	if err != nil {
		return err
	}

	return nil
}

// Run is the block's main loop. Here we listen on the different channels we set up.
func (e *FromEmail) Run() {
	var err error
	for {
		err = nil
		select {
		case msgI := <-e.inrule:
			// get id/pw/host/mailbox for IMAP
			err = e.parseAuthRules(msgI)
			if err != nil {
				e.Error(err.Error())
				continue
			}

			// if we've already got a client, close it. We need to kill it and pick up new creds.
			if e.client != nil {
				// if we're idling, term it before closing
				if e.idling {
					_, err = e.client.IdleTerm()
					if err != nil {
						// dont continue. we want to init with new creds
						e.Error(err.Error())
					}
				}
				_, err = e.client.Close(true)
				if err != nil {
					// dont continue. we want to init with new creds
					e.Error(err.Error())
				}
			}

			// initiate IMAP client with new creds
			err = e.initClient()
			if err != nil {
				e.Error(err.Error())
				continue
			}

			// do initial initial fetch on all existing unread messages
			err = e.fetchUnread()
			if err != nil {
				e.Error(err.Error())
				continue
			}

			// kick off idle in a goroutine
			go e.idle()

		case <-e.quit:
			if e.client != nil {
				// attempt to term the idle if its running
				if e.idling {
					_, err = e.client.IdleTerm()
					if err != nil {
						e.Error(err.Error())
					}
				}
				// close the IMAP conn
				_, err = e.client.Close(true)
				if err != nil {
					e.Error(err.Error())
				}
			}
			return
		case MsgChan := <-e.queryrule:
			// deal with a query request
			MsgChan <- map[string]interface{}{
				"Host":     e.host,
				"Username": e.username,
				"Password": e.password,
				"Mailbox":  e.mailbox,
			}
		}
	}
}
