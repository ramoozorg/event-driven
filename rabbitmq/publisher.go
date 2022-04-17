package rabbitmq

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/streadway/amqp"
	"go.mongodb.org/mongo-driver/bson"
	"time"
)

//Publish publishes a request to the amqp queue
func (c *Connection) Publish(exchange, routingKey string, publishOptions PublishingOptions, body interface{}) error {
	if !checkElementInSlice(c.exchanges, exchange) {
		return EXHCNAGE_NOT_FOUND_ERROR
	}
	// serialized event to bson
	b, err := bson.Marshal(body)
	if err != nil {
		return err
	}
	// try to publish event
	if err := publish(c, exchange, routingKey, publishOptions, b); err != nil {
		if errors.Is(err, CONNECTION_CLOSED_ERROR) {
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

func publish(conn *Connection, exchange, routingKey string, publishingOptions PublishingOptions, body []byte) error {
	if !conn.isConnected {
		return CONNECTION_CLOSED_ERROR
	}
	p := amqp.Publishing{
		Headers:         amqp.Table(publishingOptions.Headers),
		ContentType:     publishingOptions.ContentType,
		ContentEncoding: publishingOptions.ContentEncoding,
		DeliveryMode:    publishingOptions.DeliveryMode,
		Priority:        publishingOptions.Priority,
		CorrelationId:   publishingOptions.CorrelationId,
		ReplyTo:         publishingOptions.ReplyTo,
		Expiration:      publishingOptions.Expiration,
		MessageId:       publishingOptions.MessageId,
		Timestamp:       publishingOptions.Timestamp,
		UserId:          publishingOptions.UserId,
		AppId:           publishingOptions.AppId,
		Body:            body,
	}
	if err := conn.channel.Publish(exchange, routingKey, false, false, p); err != nil {
		return fmt.Errorf("error in Publishing: %s", err)
	}
	return nil
}
