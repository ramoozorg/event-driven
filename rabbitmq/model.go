package rabbitmq

import "github.com/streadway/amqp"

//message is the amqp request to publish
type message struct {
	RoutingKey    string
	ReplyTo       string
	ContentType   string
	CorrelationID string
	Priority      uint8
	Body          []byte
}

type delivery amqp.Delivery
