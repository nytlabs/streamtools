package library

import (
	"encoding/json"
	"net"
	"sync"

	"github.com/nytlabs/streamtools/st/blocks"
	"github.com/nytlabs/streamtools/st/util"
)

const (
	MAX_UDP_MESSAGE_SIZE = 1024
)

type listenerUDP struct {
	block   blocks.BlockInterface
	out     chan []byte
	udpConn net.PacketConn
	wait    sync.WaitGroup
	closed  bool
}

func NewListenerUDP(block blocks.BlockInterface, connectionString string, out chan []byte) (*listenerUDP, error) {
	l := &listenerUDP{
		block: block,
		out:   out,
	}

	// Try to open a new UDP connection, returning any error.
	if conn, err := net.ListenPacket("udp", connectionString); err != nil {
		return l, err
	} else {
		l.udpConn = conn
	}

	// Start the listener.
	l.wait.Add(1)
	go l.listen()

	return l, nil
}

func (l *listenerUDP) Close() {

	// Signal the listener loop to exit and close the UDP connection.
	l.closed = true
	l.udpConn.Close()

	// Wait for the listener loop to exit.
	l.wait.Wait()
}

func (l *listenerUDP) listen() {

	// Defer notification that the listener is done.
	defer l.wait.Done()

	// Create a byte buffer
	buffer := make([]byte, MAX_UDP_MESSAGE_SIZE)

	// Loop continuously.
	for !l.closed {

		// Try to read from the connection, and log the error if there is one.
		if bytes, _, err := l.udpConn.ReadFrom(buffer); err != nil {

			// Log the error.
			l.block.Error(err)
		} else {

			// Copy the message from the buffer.
			message := make([]byte, bytes)
			copy(message, buffer)

			// Dump the message onto the listener chanel.
			l.out <- message
		}
	}
}

// specify those channels we're going to use to communicate with streamtools
type FromUDP struct {
	blocks.Block
	queryrule        chan blocks.MsgChan
	inrule           blocks.MsgChan
	inpoll           blocks.MsgChan
	in               blocks.MsgChan
	out              blocks.MsgChan
	quit             blocks.MsgChan
	connectionString string
	listener         *listenerUDP
	listenerLock     sync.RWMutex
	listenerChan     chan []byte
}

// we need to build a simple factory so that streamtools can make new blocks of this kind
func NewFromUDP() blocks.BlockInterface {
	return &FromUDP{}
}

// Setup is called once before running the block. We build up the channels and
// specify what kind of block this is.
func (u *FromUDP) Setup() {
	u.Kind = "Network I/O"
	u.Desc = "listens for messages sent over UDP, emitting each into streamtools"
	u.inrule = u.InRoute("rule")
	u.queryrule = u.QueryRoute("rule")
	u.quit = u.Quit()
	u.out = u.Broadcast()
	u.listenerChan = make(chan []byte)
}

// Run is the block's main loop. Here we listen on the different channels we
// set up.
func (u *FromUDP) Run() {
	var ConnectionString string

	for {
		select {

		// Handle a rule change.
		case msgI := <-u.inrule:

			// Check for a new connection string.
			if cs, err := util.ParseString(msgI, "ConnectionString"); err != nil {
				u.Error(err)
				break
			} else {
				ConnectionString = cs
			}

			// Get the listener lock for writing.
			u.listenerLock.Lock()

			// Check if the connection string has been modified.
			if u.connectionString != ConnectionString {

				// Save the new connection string.
				u.connectionString = ConnectionString

				// Close any existing connection.
				if u.listener != nil {
					u.listener.Close()
					u.listener = nil
				}

				// Try to get a new connection.
				if l, err := NewListenerUDP(u, ConnectionString, u.listenerChan); err != nil {
					u.Error(err)
				} else {
					u.listener = l
				}
			}

			// Release the listener lock.
			u.listenerLock.Unlock()

		// Recieving a message from the listener. This is the same as from SQS
		// etc.
		case msg := <-u.listenerChan:
			var outMsg interface{}
			if err := json.Unmarshal(msg, &outMsg); err != nil {
				u.Error(err)
			} else {
				u.out <- outMsg
			}

		// Respond to a rule query.
		case MsgChan := <-u.queryrule:

			// Get the listener lock for reading.
			u.listenerLock.RLock()

			MsgChan <- map[string]interface{}{
				"ConnectionString": u.connectionString,
			}

			// Release the listener lock.
			u.listenerLock.RUnlock()

		// Shutdown everything.
		case <-u.quit:

			// Get the listener lock for writing and defer its closing.
			u.listenerLock.Lock()
			defer u.listenerLock.Unlock()

			// Clean up the listener if it exists.
			if u.listener != nil {
				u.listener.Close()
				u.listener = nil
			}

			// quit the block
			return
		}
	}
}
