package rabbitmq

import (
	"time"
)

//Consume consumes the messages from the queues and passes it as map of chan amqp.Delivery
func (c *Connection) Consume() error {
	for {
		if c.isConnected {
			break
		}
		time.Sleep(1 * time.Second)
	}
	for queue, handler := range c.queues {
		deliveries, err := c.channel.Consume(queue,
			"",
			c.ConnOpt.AutoAck,
			c.ConnOpt.ExclusiveQueue,
			false,
			c.ConnOpt.NoWait,
			nil)
		if err != nil {
			return err
		}
		if err := handler(deliveries); err != nil {
			return err
		}
	}
	return nil
}
