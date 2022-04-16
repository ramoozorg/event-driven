package rabbitmq

import (
	"fmt"
	"github.com/streadway/amqp"
	"log"
	"os"
	"time"
)

const (
	DIRECT  Kind = iota // DIRECT a message goes to the queues whose binding key exactly matches the routing key of the message.
	FANOUT              // FANOUT exchanges can be useful when the same message needs to be sent to one or more queues with consumers who may process the same message in different ways.
	TOPIC               // TOPIC exchange is similar to direct exchange, but the routing is done according to the routing pattern. Instead of using fixed routing key, it uses wildcards.
	HEADERS             // HEADERS exchange routes messages based on arguments containing headers and optional values. It uses the message header attributes for routing.
)

const (
	delayReconnectTime = 5 * time.Second
)

// NewConnection create a rabbitmq connection object
func NewConnection(serviceName string, options *Options, done chan os.Signal) (*Connection, error) {
	opts, err := validateOptions(serviceName, options)
	if err != nil {
		return nil, err
	}
	connObj := &Connection{
		ServiceCallerName: serviceName,
		ConnOpt:           opts,
		done:              done,
		alive:             true,
	}
	go connObj.handleReconnect(opts.UriAddress)
	return connObj, nil
}

// NewEncodedConn will wrap an existing Connection and utilize the appropriate registered encoder
func NewEncodedConn(c *Connection, encType string) (*EncodedConn, error) {
	if c == nil {
		return nil, NIL_CCONECTION_ERROR
	}
	ec := &EncodedConn{Conn: c, Enc: EncoderForType(encType)}
	if ec.Enc == nil {
		return nil, fmt.Errorf("no encoder registered for '%s'", encType)
	}
	return ec, nil
}

// connect dial to rabbitMQ server and declare exchange
func (c *Connection) connect() bool {
	conn, err := amqp.Dial(c.ConnOpt.UriAddress)
	if err != nil {
		return false
	}
	ch, err := conn.Channel()
	if err != nil {
		return false
	}
	c.updateConnection(conn, ch)
	c.isConnected = true
	return true
}

// handleReconnect if closing rabbitMQ try to connect rabbitMQ continuously
func (c *Connection) handleReconnect(addr string) {
	for c.alive {
		c.isConnected = false
		now := time.Now()
		log.Printf("attempting to connect to rabbitMQ %v", addr)
		retryCount := 0
		for !c.connect() {
			if !c.alive {
				return
			}
			select {
			case <-c.done:
				return
			case <-time.After(delayReconnectTime + time.Duration(retryCount)*time.Second):
				log.Printf("cannot connect to rabbitMQ try connecting to rabbitMQ...")
				if retryCount != 10 {
					retryCount++
				}
			}
		}
		log.Printf("connected to rabbitMQ after %v second", time.Since(now).Seconds())
		select {
		case <-c.done:
			return
		case <-c.notifyClose:
		}
	}
}

// updateConnection update connection and channel in memory
func (c *Connection) updateConnection(connection *amqp.Connection, channel *amqp.Channel) {
	c.conn = connection
	c.channel = channel
	c.notifyClose = make(chan *amqp.Error)
	c.channel.NotifyClose(c.notifyClose)
}

// ExchangeDeclare declare new exchange with specific kind (direct, topic, fanout, headers)
func (c *Connection) ExchangeDeclare(exchange string, kind Kind) error {
	if checkElementInSlice(c.exchanges, exchange) {
		return EXCHANGE_ALREADY_EXISTS_ERROR
	}
	c.exchanges = append(c.exchanges, exchange)
	if err := c.channel.ExchangeDeclare(
		exchange,
		kind.String(),
		c.ConnOpt.DurableExchange,
		c.ConnOpt.AutoDelete,
		false,
		c.ConnOpt.NoWait,
		nil); err != nil {
		return err
	}
	return nil
}

// QueueDeclare declare new queue and bind queue and bind exchange with routing key
func (c *Connection) QueueDeclare(queue, exchange, routingKey string, messageHandler MessageHandler) error {
	if c.queues == nil {
		c.queues = make(map[string]MessageHandler)
	}
	if _, ok := c.queues[queue]; ok {
		return QUEUE_ALREADY_EXISTS_ERROR
	} else {
		c.queues[queue] = messageHandler
	}
	if _, err := c.channel.QueueDeclare(
		queue,
		c.ConnOpt.DurableExchange,
		c.ConnOpt.AutoDelete,
		c.ConnOpt.ExclusiveQueue,
		c.ConnOpt.NoWait,
		nil,
	); err != nil {
		return err
	}

	if err := c.channel.QueueBind(
		queue,
		routingKey,
		exchange,
		c.ConnOpt.NoWait,
		nil,
	); err != nil {
		return err
	}
	return nil
}

// IsConnected check rabbitMQ client is connected
func (c *Connection) IsConnected() bool {
	return c.isConnected
}

// GetExchangeList return list of exchanges
func (c *Connection) GetExchangeList() []string {
	return c.exchanges
}

// GetQueueList return list of queues with handlers
func (c *Connection) GetQueueList() map[string]MessageHandler {
	return c.queues
}

//Close stop rabbitMQ client
func (c *Connection) Close() error {
	if !c.isConnected {
		return nil
	}
	c.alive = false
	if err := c.channel.Close(); err != nil {
		return err
	}
	if err := c.conn.Close(); err != nil {
		return err
	}
	c.isConnected = false
	log.Printf("gracefully stopped rabbitMQ connection")
	return nil
}

func validateOptions(serviceName string, newOpt *Options) (*Options, error) {
	opt := getDefaultOptions()
	if len(serviceName) == 0 {
		return nil, SERVICE_NAME_ERROR
	}
	if len(newOpt.UriAddress) == 0 {
		return nil, URI_ADDRESS_ERROR
	} else {
		opt.UriAddress = newOpt.UriAddress
	}
	if newOpt.DurableExchange != opt.DurableExchange {
		opt.DurableExchange = newOpt.DurableExchange
	}
	if newOpt.AutoAck != opt.AutoAck {
		opt.AutoAck = newOpt.AutoAck
	}
	if newOpt.AutoDelete != opt.AutoDelete {
		opt.AutoDelete = newOpt.AutoDelete
	}
	if newOpt.NoWait != opt.NoWait {
		opt.NoWait = newOpt.NoWait
	}
	if newOpt.ExclusiveQueue != opt.ExclusiveQueue {
		opt.ExclusiveQueue = newOpt.ExclusiveQueue
	}
	return opt, nil
}

func checkElementInSlice(slice []string, newElement string) bool {
	if slice != nil {
		for _, s := range slice {
			if s == newElement {
				return true
			}
		}
	}
	return false
}

// String exchange type as string
func (k Kind) String() string {
	switch k {
	case DIRECT:
		return "direct"
	case FANOUT:
		return "fanout"
	case TOPIC:
		return "topic"
	case HEADERS:
		return "headers"
	default:
		return "topic"
	}
}
