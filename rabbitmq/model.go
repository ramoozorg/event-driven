package rabbitmq

import (
	"github.com/streadway/amqp"
	"os"
)

type (
	Kind           int                  // Kind is exchange type
	MessageHandler func(Delivery) error // MessageHandler handle message from specific queue and routingKey
	Delivery       <-chan amqp.Delivery // Delivery is a channel for deliver published message
	Headers        amqp.Table           // Headers table for set message header when publishing
)

//Connection is the structure of amqp event connection
type Connection struct {
	conn              *amqp.Connection // conn rabbitMQ connection Object
	channel           *amqp.Channel    // channel amqp channel Object
	done              chan os.Signal
	notifyClose       chan *amqp.Error
	isConnected       bool
	alive             bool
	exchanges         []string                  // exchanges list
	queues            map[string]MessageHandler // queue and message handler
	ServiceCallerName string
	ConnOpt           *Options
}
