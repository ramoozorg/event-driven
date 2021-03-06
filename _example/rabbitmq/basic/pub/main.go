package main

import (
	"github.com/ramoozorg/event-driven/rabbitmq"
	"os"
)

type person struct {
	Name string `bson:"name" json:"name"`
	Age  int    `bson:"age" json:"age"`
}

func main() {
	chSignal := make(chan os.Signal)
	conn, err := rabbitmq.NewConnection("test", &rabbitmq.Options{
		UriAddress:      "amqp://guest:guest@localhost:5672",
		DurableExchange: true,
		AutoAck:         true,
		ExclusiveQueue:  false,
	}, chSignal)
	if err != nil {
		panic(err)
	}

	if err := conn.ExchangeDeclare("exchange1", rabbitmq.TOPIC); err != nil {
		panic(err)
	}
	if err := conn.DeclarePublisherQueue("queue1", "exchange1", "rk", "rk2"); err != nil {
		panic(err)
	}
	if err := conn.DeclarePublisherQueue("queue2", "exchange1", "rk3"); err != nil {
		panic(err)
	}
	if err := NewEventPublish(conn); err != nil {
		panic(err)
	}
}

func NewEventPublish(conn *rabbitmq.Connection) error {
	p := person{Name: "rs", Age: 22}
	q := person{Name: "reza", Age: 23}
	if err := conn.Publish("exchange1", "rk", p, rabbitmq.PublishingOptions{}); err != nil {
		return err
	}
	if err := conn.Publish("exchange1", "rk3", q, rabbitmq.PublishingOptions{}); err != nil {
		return err
	}
	return nil
}
