package rabbitmq

import (
	"fmt"
	"git.ramooz.org/ramooz/golang-components/logger"
	"github.com/streadway/amqp"
	"os"
	"sync"
	"time"
)

const delayReconnectTime = 5 * time.Second

//Connection is the connection and channel of amqp
type Connection struct {
	conn              *amqp.Connection
	channel           *amqp.Channel
	done              chan os.Signal
	notifyClose       chan *amqp.Error
	mu                sync.RWMutex
	isConnected       bool
	alive             bool
	logger            *logger.LogService
	ServiceCallerName string
	ConnOpt           *Options
	Exchange          []*string // Exchange name
	RoutingKeys       []*string // RoutingKeys for messages
	Queues            []*string // Queues name
	Kind              *string
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
		logger:            newLogger(),
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
		c.logger.Errorf("failed connect to rabbitMQ server %v, got error %v", c.ConnOpt.UriAddress, err)
		return false
	}
	ch, err := c.conn.Channel()
	if err != nil {
		c.logger.Errorf("failed connect to rabbitMQ channel, got error %v", err)
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
		c.logger.Infof("attempting to connect to rabbitMQ %v", addr)
		retryCount := 0
		for !c.connect() {
			if !c.alive {
				return
			}
			select {
			case <-c.done:
				return
			case <-time.After(delayReconnectTime + time.Duration(retryCount)*time.Second):
				c.logger.Warn("cannot connect to rabbitMQ try connecting to rabbitMQ...")
			}
		}
		c.logger.Infof("connected to rabbitMQ after %v second", time.Since(now).Seconds())
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

func newLogger() *logger.LogService {
	return logger.NewLogger(10001, "event-component", &logger.Options{Colorable: true, Development: true})
}
