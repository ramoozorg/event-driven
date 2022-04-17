package rabbitmq

import (
	"git.ramooz.org/ramooz/golang-components/logger"
	"github.com/streadway/amqp"
	"os"
	"time"
)

type (
	Kind           int                                   // Kind is exchange type
	MessageHandler func(queue string, delivery Delivery) // MessageHandler handle message from specific queue and routingKey
	Delivery       amqp.Delivery                         // Delivery is a channel for deliver published message
	Headers        amqp.Table                            // Headers table for set message header when publishing
)

// Connection is the structure of amqp event connection
type Connection struct {
	conn              *amqp.Connection // conn rabbitMQ connection Object
	channel           *amqp.Channel    // channel amqp channel Object
	done              chan os.Signal
	notifyClose       chan *amqp.Error
	isConnected       bool
	alive             bool
	exchanges         []string                  // exchanges list
	queues            map[string]MessageHandler // queue and message handler
	logger            *logger.LogService
	ServiceCallerName string
	ConnOpt           *Options
}

// PublishingOptions options for message
type PublishingOptions struct {
	Headers Headers // rabbitMQ message headers
	// Properties
	ContentType     string    // MIME content type
	ContentEncoding string    // MIME content encoding
	DeliveryMode    uint8     // Transient (0 or 1) or Persistent (2)
	Priority        uint8     // 0 to 9
	CorrelationId   string    // correlation identifier
	ReplyTo         string    // address to to reply to (ex: RPC)
	Expiration      string    // message expiration spec
	MessageId       string    // message identifier
	Timestamp       time.Time // message timestamp
	Type            string    // message type name
	UserId          string    // creating user id - ex: "guest"
	AppId           string    // creating application id
}
