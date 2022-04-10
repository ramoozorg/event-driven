package main

import (
	"git.ramooz.org/ramooz/golang-components/event-driven/rabbitmq"
	"log"
)

type Person struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func main() {
	conn, err := rabbitmq.NewConnection("test", &rabbitmq.Options{
		UriAddress:      "amqp://guest:guest@localhost:5672",
		Exchange:        "test",
		RoutingKeys:     []string{"RoutingA", "RoutingB"},
		Queues:          []string{"QueueA"},
		DurableExchange: true,
		AutoAck:         true,
		ExclusiveQueue:  false,
	})
	if err != nil {
		panic(err)
	}
	if err := conn.Connect(); err != nil {
		panic(err)
	}
	if err := conn.BindQueue(); err != nil {
		panic(err)
	}
	if err := NewMessagePublish(conn); err != nil {
		panic(err)
	}
}

func NewMessagePublish(conn *rabbitmq.Connection) error {
	for _, key := range conn.ConnOpt.RoutingKeys {
		msg := "hello world"
		if err := conn.Publish(key, "", "", 0, msg); err != nil {
			return err
		}
		log.Println("message published ", msg)
	}
	return nil
}
