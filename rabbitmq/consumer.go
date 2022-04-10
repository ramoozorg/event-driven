package rabbitmq

import (
	"github.com/streadway/amqp"
)

//Consume consumes the messages from the queues and passes it as map of chan of amqp.Delivery
func (c *Connection) Consume() (map[string]<-chan amqp.Delivery, error) {
	msg := make(map[string]<-chan amqp.Delivery)
	for _, queue := range c.ConnOpt.Queues {
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

// HandleConsumedDeliveries handles the consumed deliveries from the queues. Should be called only for a consumer connection
func (c *Connection) HandleConsumedDeliveries(key string, delivery <-chan amqp.Delivery, fn func(*Connection, <-chan amqp.Delivery)) {
	for {
		go fn(c, delivery)
		if err := <-c.err; err != nil {
			c.Reconnect()
			deliveries, err := c.Consume()
			if err != nil {
				panic(err)
			}
			delivery = deliveries[key]
		}
	}
}

// HandleConsumedDeliveries handles the consumed deliveries from the queues and decode to specific structure. Should be called only for a consumer connection
func (e *EncodedConn) HandleConsumedDeliveries(key string, msgStructure interface{}, delivery <-chan amqp.Delivery, fn func(*EncodedConn, <-chan amqp.Delivery)) {
	//TODO: we need msgStructure decoded for body
	for {
		go fn(e, delivery)
		if err := e.Conn.err; err != nil {
			e.Conn.Reconnect()
			encDeliveries := map[string]<-chan amqp.Delivery{}
			deliveries, err := e.Conn.Consume()
			if err != nil {
				panic(err)
			}
			for _, msg := range deliveries {
				inChMsg := <-msg
				e.Enc.Decode(inChMsg.Body, msgStructure)
				outChMsg := make(chan amqp.Delivery)
				outChMsg <- inChMsg
				encDeliveries[key] = outChMsg
			}
			delivery = encDeliveries[key]
		}
	}
}
