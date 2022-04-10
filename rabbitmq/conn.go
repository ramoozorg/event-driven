package rabbitmq

import (
	"fmt"
	"github.com/streadway/amqp"
	"sync"
)

//Connection is the connection and channel of amqp
type Connection struct {
	ServiceCallerName string
	conn              *amqp.Connection
	channel           *amqp.Channel
	ConnOpt           *Options
	mu                sync.RWMutex
	err               chan error
}

// NewConnection create a rabbitmq connection object
func NewConnection(serviceName string, options *Options) (*Connection, error) {
	opts, err := validateOptions(serviceName, options)
	if err != nil {
		return nil, err
	}
	return &Connection{
		ServiceCallerName: serviceName,
		ConnOpt:           opts,
		err:               make(chan error),
	}, nil
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

// Connect dial to rabbitMQ server and declare exchange
func (c *Connection) Connect() error {
	var err error
	c.conn, err = amqp.Dial(c.ConnOpt.UriAddress)
	if err != nil {
		return err
	}
	//Listen to NotifyClose from connection
	go func() {
		<-c.conn.NotifyClose(make(chan *amqp.Error))
		c.err <- CONNECTION_CLOSED_ERROR
	}()
	c.channel, err = c.conn.Channel()
	if err != nil {
		return err
	}
	if err := c.channel.ExchangeDeclare(
		c.ConnOpt.Exchange,
		c.ConnOpt.Kind,
		c.ConnOpt.DurableExchange,
		c.ConnOpt.AutoDelete,
		false,
		c.ConnOpt.NoWait,
		nil,
	); err != nil {
		return fmt.Errorf("error in exchange declare %v", err)
	}
	return nil
}

// BindQueue declare and bind new queues
func (c *Connection) BindQueue() error {
	for _, queue := range c.ConnOpt.Queues {
		if _, err := c.channel.QueueDeclare(queue,
			c.ConnOpt.DurableExchange,
			c.ConnOpt.AutoDelete,
			c.ConnOpt.ExclusiveQueue,
			c.ConnOpt.NoWait, nil); err != nil {
			return fmt.Errorf("queue declare error %v", err)
		}
		for _, key := range c.ConnOpt.RoutingKeys {
			if err := c.channel.QueueBind(queue,
				key,
				c.ConnOpt.Exchange,
				c.ConnOpt.NoWait,
				nil); err != nil {
				return fmt.Errorf("queue bind error %v", err)
			}
		}
	}
	return nil
}

// Reconnect automatic reconnect publisher and consumer if connection dropped
func (c *Connection) Reconnect() error {
	if err := c.Connect(); err != nil {
		return err
	}
	if err := c.BindQueue(); err != nil {
		return err
	}
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
	if len(newOpt.Exchange) != 0 {
		opt.Exchange = newOpt.Exchange
	}
	if len(newOpt.RoutingKeys) == 0 {
		return nil, ROUTING_KEYS_EMPTY_ERROR
	} else {
		opt.RoutingKeys = newOpt.RoutingKeys
	}
	if len(newOpt.Queues) != 0 {
		opt.Queues = newOpt.Queues
	}
	if len(newOpt.Kind) != 0 {
		opt.Kind = newOpt.Kind
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
