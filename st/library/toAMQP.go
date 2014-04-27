package library

import (
	"encoding/json"
	"github.com/nytlabs/streamtools/st/blocks"
	"github.com/nytlabs/streamtools/st/util"
	"github.com/streadway/amqp"
)

// specify those channels we're going to use to communicate with streamtools
type ToAMQP struct {
	blocks.Block
	queryrule chan blocks.MsgChan
	in        blocks.MsgChan
	inrule    blocks.MsgChan
	quit      blocks.MsgChan

	host          string
	port          string
	username      string
	password      string
	exchange      string
	exchange_type string
	routingkey    string
}

// a bit of boilerplate for streamtools
func NewToAMQP() blocks.BlockInterface {
	return &ToAMQP{host: "localhost", username: "guest",
		password: "guest", exchange: "amq.topic", routingkey: "streamtools",
		port: "5672", exchange_type: "topic"}
}

func (b *ToAMQP) Setup() {
	b.Kind = "ToAMQP"
	b.Desc = "send messages to an exchange on an AMQP broker"
	b.in = b.InRoute("in")
	b.inrule = b.InRoute("rule")
	b.queryrule = b.QueryRoute("rule")
	b.quit = b.Quit()
}

// connects to an AMQP topic and emits each message into streamtools.
func (b *ToAMQP) Run() {
	var err error
	var conn *amqp.Connection
	var amqp_chan *amqp.Channel

	// Pickup defaults from construction
	host := b.host
	port := b.port
	username := b.username
	password := b.password
	routingkey := b.routingkey
	exchange := b.exchange
	exchange_type := b.exchange_type

	for {
		select {
		case ruleI := <-b.inrule:

			routingkey, err = util.ParseString(ruleI, "RoutingKey")
			if err != nil {
				b.Error(err)
				continue
			}
			exchange, err = util.ParseString(ruleI, "Exchange")
			if err != nil {
				b.Error(err)
				continue
			}
			exchange_type, err = util.ParseString(ruleI, "ExchangeType")
			if err != nil {
				b.Error(err)
				continue
			}
			host, err = util.ParseString(ruleI, "Host")
			if err != nil {
				b.Error(err)
				continue
			}
			port, err = util.ParseString(ruleI, "Port")
			if err != nil {
				b.Error(err)
				continue
			}
			username, err = util.ParseString(ruleI, "Username")
			if err != nil {
				b.Error(err)
				continue
			}
			password, err = util.ParseString(ruleI, "Password")
			if err != nil {
				b.Error(err)
				continue
			}

			conn, err = amqp.Dial("amqp://" + username + ":" + password + "@" + host + ":" + port + "/")
			if err != nil {
				b.Error(err)
				continue
			}

			amqp_chan, err = conn.Channel()
			if err != nil {
				b.Error(err)
				continue
			}

			err = amqp_chan.ExchangeDeclare(
				exchange,      // name
				exchange_type, // type
				true,          // durable
				false,         // auto-deleted
				false,         // internal
				false,         // noWait
				nil,           // arguments
			)
			if err != nil {
				b.Error(err)
				continue
			}

		case msg := <-b.in:
			if conn == nil || amqp_chan == nil {
				continue
			}

			msgBytes, err := json.Marshal(msg)
			if err != nil {
				b.Error(err)
			}
			if len(msgBytes) == 0 {
				continue
			}

			err = amqp_chan.Publish(
				exchange,
				routingkey,
				false,
				false,
				amqp.Publishing{
					Headers:         amqp.Table{},
					ContentType:     "text/plain",
					ContentEncoding: "",
					Body:            msgBytes,
					DeliveryMode:    amqp.Transient,
					Priority:        0,
				},
			)
			if err != nil {
				b.Error(err)
				continue
			}
		case <-b.quit:
			if amqp_chan != nil {
				amqp_chan.Close()
			}
			if conn != nil {
				conn.Close()
			}
			return
		case c := <-b.queryrule:
			c <- map[string]interface{}{
				"Host":         host,
				"Port":         port,
				"Username":     username,
				"Password":     password,
				"Exchange":     exchange,
				"ExchangeType": exchange_type,
				"RoutingKey":   routingkey,
			}
		}
	}
}
