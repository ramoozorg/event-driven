package rabbitmq

import (
	"fmt"
	"github.com/streadway/amqp"
	"log"
	"os"
	"time"
)

const delayReconnectTime = 5 * time.Second

//Connection is the connection and channel of amqp
type Connection struct {
	conn              *amqp.Connection
	channel           *amqp.Channel
	done              chan os.Signal
	notifyClose       chan *amqp.Error
	isConnected       bool
	alive             bool
	exchanges         map[string]string // exchanges list
	routingKeys       []string          // routingKeys for messages
	queues            []string          // queues name
	ServiceCallerName string
	ConnOpt           *Options
}

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
	c.exchanges = make(map[string]string)
	c.notifyClose = make(chan *amqp.Error)
	c.channel.NotifyClose(c.notifyClose)
}

// ExchangeDeclare declare new exchange with specific kind (direct, topic and ...),
// exchanges map[exchangeName]kind
func (c *Connection) ExchangeDeclare(exchanges map[string]string) error {
	for exchange, kind := range exchanges {
		if c.exchanges == nil {
			c.exchanges = make(map[string]string)
		}
		c.exchanges[exchange] = kind
		if err := c.channel.ExchangeDeclarePassive(
			exchange,
			kind,
			c.ConnOpt.DurableExchange,
			c.ConnOpt.AutoDelete,
			false,
			c.ConnOpt.NoWait,
			nil); err != nil {
			return err
		}
	}
	return nil
}

// QueueDeclare declare new queue for rabbitMQ
func (c *Connection) QueueDeclare(queues []string) error {
	for _, queue := range queues {
		if !checkExistsInSlice(c.queues, queue) {
			c.queues = append(c.queues, queue)
		} else {
			return fmt.Errorf("queue %v already exists", queue)
		}
		if _, err := c.channel.QueueDeclarePassive(
			queue,
			c.ConnOpt.DurableExchange,
			c.ConnOpt.AutoDelete,
			c.ConnOpt.ExclusiveQueue,
			c.ConnOpt.NoWait,
			nil,
		); err != nil {
			return err
		}
	}
	return nil
}

// RoutingKeyDeclare declare routing keys
func (c *Connection) RoutingKeyDeclare(routingKeys []string) error {
	for _, key := range routingKeys {
		if !checkExistsInSlice(c.routingKeys, key) {
			c.routingKeys = append(c.routingKeys, key)
		} else {
			return fmt.Errorf("routing key %v already exists", key)
		}
	}
	return nil
}

func (c *Connection) QueueBind() error {
	if c.queues == nil {
		return fmt.Errorf("queues is empty, please first declare queues")
	}
	if c.exchanges == nil {
		return fmt.Errorf("exchanges is empty, please first declare exchanges")
	}
	if c.routingKeys == nil {
		return fmt.Errorf("routingKeys is empty, please first declare routingKeys")
	}
	for exchange := range c.exchanges {
		for _, queue := range c.queues {
			for _, key := range c.routingKeys {
				if err := c.channel.QueueBind(
					queue,
					key,
					exchange,
					c.ConnOpt.NoWait,
					nil,
				); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// IsConnected check rabbitMQ client is connect
func (c *Connection) IsConnected() bool {
	return c.isConnected
}

// GetExchangeList return list of exchanges
func (c *Connection) GetExchangeList() map[string]string {
	return c.exchanges
}

// GetQueueList return list of queues
func (c *Connection) GetQueueList() []string {
	return c.queues
}

// GetRoutingKeys return list of RoutingKeys
func (c *Connection) GetRoutingKeys() []string {
	return c.routingKeys
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

// IsEmpty check message is empty
func (m Message) IsEmpty() bool {
	if len(m.Body) == 0 {
		return true
	}
	return false
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

func checkExistsInSlice(slice []string, newElement string) bool {
	if slice != nil {
		for _, s := range slice {
			if s == newElement {
				return true
			}
		}
	}
	return false
}
