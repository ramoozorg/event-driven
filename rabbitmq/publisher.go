package rabbitmq

import (
	"fmt"
	"github.com/streadway/amqp"
	"go.mongodb.org/mongo-driver/bson"
)

//Publish publishes a request to the amqp queue
func (c *Connection) Publish(key, replayTo, contentType, correlationID string, priority uint8, body interface{}) error {
	msg := message{RoutingKey: key, ReplyTo: replayTo, ContentType: contentType, CorrelationID: correlationID, Priority: priority}
	b, err := bson.Marshal(body)
	if err != nil {
		return err
	}
	msg.Body = b
	return publish(msg, c)
}

//// Publish publishes the data argument to the push amqp message. The data argument
//// will be encoded using the associated encoder.
//func (c *EncodedConn) Publish(key, replayTo, contentType, correlationID string, priority uint8, body interface{}) error {
//	msg := message{RoutingKey: key, ReplyTo: replayTo, ContentType: contentType, CorrelationID: correlationID, Priority: priority}
//	encBody, err := c.Enc.Encode(body)
//	if err != nil {
//		return err
//	}
//	msg.Body = encBody
//	return publish(msg, c.Conn)
//}

func publish(msg message, conn *Connection) error {
	if !conn.isConnected {
		return CONNECTION_CLOSED_ERROR
	}
	p := amqp.Publishing{
		ContentType:   msg.ContentType,
		CorrelationId: msg.CorrelationID,
		Body:          msg.Body,
		ReplyTo:       msg.ReplyTo,
	}
	for exchange, _ := range conn.exchanges {
		if err := conn.channel.Publish(exchange, msg.RoutingKey, false, false, p); err != nil {
			return fmt.Errorf("Error in Publishing: %s", err)
		}
	}
	return nil
}
