package library

import (
	"encoding/json"

	"github.com/nytlabs/streamtools/st/blocks"
	"github.com/nytlabs/streamtools/st/util"
	"github.com/streadway/amqp"
)

// specify those channels we're going to use to communicate with streamtools
type FromAMQP struct {
	blocks.Block
	queryrule chan blocks.MsgChan
	inrule    blocks.MsgChan
	out       blocks.MsgChan
	quit      blocks.MsgChan
}

// a bit of boilerplate for streamtools
func NewFromAMQP() blocks.BlockInterface {
	return &FromAMQP{}
}

func (b *FromAMQP) Setup() {
	b.Kind = "Queues"
	b.Desc = "reads from a topic on AMQP broker as specified in this block's rules"
	b.inrule = b.InRoute("rule")
	b.queryrule = b.QueryRoute("rule")
	b.quit = b.Quit()
	b.out = b.Broadcast()
}

type readWriteAMQPHandler struct {
	toOut   blocks.MsgChan
	toError chan error
}

func (self readWriteAMQPHandler) handle(deliveries <-chan amqp.Delivery) {
	for d := range deliveries {
		var msg interface{}
		s := string(d.Body[:])
		err := json.Unmarshal(d.Body, &msg)
		if err != nil {
			msg = map[string]interface{}{
				"data": s,
			}
		}
		self.toOut <- msg
	}
}

// connects to an AMQP topic and emits each message into streamtools.
func (b *FromAMQP) Run() {
	var err error
	var conn *amqp.Connection
	var amqp_chan *amqp.Channel
	toOut := make(blocks.MsgChan)
	toError := make(chan error)

	host := "localhost"
	port := "5672"
	username := "guest"
	password := "guest"
	routingkey := "#"
	exchange := "amq.topic"
	exchange_type := "topic"

	for {
		select {
		case msg := <-toOut:
			b.out <- msg
		case err := <-toError:
			b.Error(err)
		case ruleI := <-b.inrule:
			rule := ruleI.(map[string]interface{})

			routingkey, err = util.ParseString(rule, "RoutingKey")
			if err != nil {
				b.Error(err)
				continue
			}
			exchange, err = util.ParseString(rule, "Exchange")
			if err != nil {
				b.Error(err)
				continue
			}
			exchange_type, err = util.ParseString(rule, "ExchangeType")
			if err != nil {
				b.Error(err)
				continue
			}
			host, err = util.ParseString(rule, "Host")
			if err != nil {
				b.Error(err)
				continue
			}
			port, err = util.ParseString(rule, "Port")
			if err != nil {
				b.Error(err)
				continue
			}
			username, err = util.ParseString(rule, "Username")
			if err != nil {
				b.Error(err)
				continue
			}
			password, err = util.ParseString(rule, "Password")
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

			queue, err := amqp_chan.QueueDeclare(
				"",    // name
				false, // durable
				true,  // delete when unused
				false, // exclusive
				false, // noWait
				nil,   // arguments
			)
			if err != nil {
				b.Error(err)
				continue
			}

			err = amqp_chan.QueueBind(
				queue.Name, // queue name
				routingkey, // routing key
				exchange,   // exchange
				false,
				nil,
			)

			if err != nil {
				b.Error(err)
				continue
			}

			deliveries, err := amqp_chan.Consume(
				queue.Name, // name
				"",         // consumerTag
				true,       // noAck
				false,      // exclusive
				false,      // noLocal
				false,      // noWait
				nil,        // arguments
			)
			if err != nil {
				b.Error(err)
				continue
			}

			h := readWriteAMQPHandler{toOut, toError}
			go h.handle(deliveries)
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
