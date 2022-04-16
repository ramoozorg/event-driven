package rabbitmq

import (
	"fmt"
	"github.com/streadway/amqp"
	"go.mongodb.org/mongo-driver/bson"
	"time"
)

//Publish publishes a request to the amqp queue
func (c *Connection) Publish(exchange, routingKey, replayTo, contentType, correlationID string, headers amqp.Table, body interface{}) error {
	if !checkElementInSlice(c.exchanges, exchange) {
		return EXHCNAGE_NOT_FOUND_ERROR
	}
	// serialized message to bson
	b, err := bson.Marshal(body)
	if err != nil {
		return err
	}
	// try to publish event
	if err := publish(exchange, routingKey, amqp.Publishing{
		Headers:       headers,
		ContentType:   contentType,
		CorrelationId: correlationID,
		ReplyTo:       replayTo,
		Body:          b,
	}, c); err != nil {
		if err == CONNECTION_CLOSED_ERROR {
			for {
				if c.isConnected {
					break
				}
				time.Sleep(1 * time.Second)
			}
		}
		return err
	}
	return nil
}

func publish(exchange, routingKey string, msg amqp.Publishing, conn *Connection) error {
	if !conn.isConnected {
		return CONNECTION_CLOSED_ERROR
	}
	p := amqp.Publishing{
		ContentType:   msg.ContentType,
		CorrelationId: msg.CorrelationId,
		Body:          msg.Body,
		ReplyTo:       msg.ReplyTo,
	}
	if err := conn.channel.Publish(exchange, routingKey, false, false, p); err != nil {
		return fmt.Errorf("error in Publishing: %s", err)
	}
	return nil
}
