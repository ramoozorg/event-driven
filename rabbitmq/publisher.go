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
	return publish(msg, c)
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
	return publish(msg, c.Conn)
}

func publish(msg message, conn *Connection) error {
	select {
	case err := <-conn.err:
		if err != nil {
			conn.Reconnect()
		}
	default:
	}
	p := amqp.Publishing{
		ContentType:   msg.ContentType,
		CorrelationId: msg.CorrelationID,
		Body:          msg.Body,
		ReplyTo:       msg.ReplyTo,
	}
	if err := conn.channel.Publish(conn.ConnOpt.Exchange, msg.RoutingKey, false, false, p); err != nil {
		return fmt.Errorf("Error in Publishing: %s", err)
	}
	return nil
}
