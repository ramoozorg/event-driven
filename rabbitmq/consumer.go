package rabbitmq

import (
	"github.com/streadway/amqp"
	"time"
)

//Consume consumes the messages from the queues and passes it as map of chan amqp.Delivery
func (c *Connection) Consume() (map[string]<-chan amqp.Delivery, error) {
	for {
		if c.isConnected {
			break
		}
		time.Sleep(1 * time.Second)
	}
	msg := make(map[string]<-chan amqp.Delivery)
	for _, queue := range c.queues {
		deliveries, err := c.channel.Consume(queue,
			"",
			c.ConnOpt.AutoAck,
			c.ConnOpt.ExclusiveQueue,
			false,
			c.ConnOpt.NoWait,
			nil)
		if err != nil {
			return nil, err
		}
		msg[queue] = deliveries
	}
	return msg, nil
}

// MessageHandler create new Message structure for received new message
func (c *Connection) MessageHandler(queue string, delivery <-chan amqp.Delivery) Message {
	select {
	case msg, _ := <-delivery:
		return Message{
			Exchange:   msg.Exchange,
			Queue:      queue,
			RoutingKey: msg.RoutingKey,
			Body:       msg.Body,
		}
	}
}
