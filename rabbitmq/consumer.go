package rabbitmq

import (
	"time"
)

//Consume consumes the messages from the queues and passes it as map of chan amqp.Delivery
func (c *Connection) Consume() error {
	for queue, handler := range c.queues {
		go c.consume(queue, handler)
	}
	return nil
}

func (c *Connection) consume(queue string, messageHandler MessageHandler) {
	deliveries, _ := c.channel.Consume(queue,
		"",
		c.ConnOpt.AutoAck,
		c.ConnOpt.ExclusiveQueue,
		false,
		c.ConnOpt.NoWait,
		nil)
	for {
		select {
		case msg := <-deliveries:
			if len(msg.Body) != 0 {
				messageHandler(queue, Delivery(msg))
			}
		case <-c.notifyClose:
			for {
				if c.isConnected {
					c.consume(queue, messageHandler)
					break
				}
				time.Sleep(1 * time.Second)
			}
		default:
		}
	}
}
