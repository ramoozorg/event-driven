package rabbitmq

import (
	"log"
	"time"
)

//Consume consumes the events from the queues and passes it as map of chan amqp.Delivery
func (c *Connection) Consume() error {
	for queue, handler := range c.queues {
		queue := queue
		handler := handler
		go func() {
			c.consume(queue, handler)
			if r := recover(); r != nil {
				log.Printf("consumer got panic %v and recovered", r)
			}
		}()
	}
	return nil
}

func (c *Connection) consume(queue string, eventHandler EventHandler) {
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
				eventHandler(queue, Delivery(msg))
			}
		case <-c.notifyClose:
			for {
				if c.isConnected {
					c.consume(queue, eventHandler)
					break
				}
				time.Sleep(1 * time.Second)
			}
		}
	}
}
