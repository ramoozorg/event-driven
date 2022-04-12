package rabbitmq

import (
	"fmt"
	"github.com/streadway/amqp"
	"go.mongodb.org/mongo-driver/bson"
)

//Publish publishes a request to the amqp queue
func (c *Connection) Publish(key, replayTo, correlationID string, priority uint8, body interface{}) error {
	msg := message{RoutingKey: key, ReplyTo: replayTo, ContentType: "text/plain", CorrelationID: correlationID, Priority: priority}
	b, err := bson.Marshal(body)
	if err != nil {
		return err
	}
	msg.Body = b
	return c.publish(msg)
}

// Publish publishes the data argument to the push amqp message. The data argument
// will be encoded using the associated encoder.
func (c *EncodedConn) Publish(key, replayTo, correlationID string, priority uint8, body interface{}) error {
	msg := message{RoutingKey: key, ReplyTo: replayTo, ContentType: "text/plain", CorrelationID: correlationID, Priority: priority}
	encBody, err := c.Enc.Encode(body)
	if err != nil {
		return err
	}
	msg.Body = encBody
	return c.Conn.publish(msg)
}

func (c *Connection) publish(m message) error {
	select {
	case err := <-c.err:
		if err != nil {
			c.Reconnect()
		}
	default:
	}
	p := amqp.Publishing{
		ContentType:   m.ContentType,
		CorrelationId: m.CorrelationID,
		Body:          m.Body,
		ReplyTo:       m.ReplyTo,
	}
	if err := c.channel.Publish(c.ConnOpt.Exchange, m.RoutingKey, false, false, p); err != nil {
		return fmt.Errorf("Error in Publishing: %s", err)
	}
	return nil
}
