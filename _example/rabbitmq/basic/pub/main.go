package main

import (
	"git.ramooz.org/ramooz/golang-components/event-driven/rabbitmq"
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
	if err := conn.DeclarePublisherQueue("queue1", "exchange1", "rk"); err != nil {
		panic(err)
	}
	if err := conn.DeclarePublisherQueue("queue2", "exchange1", "rk2"); err != nil {
		panic(err)
	}
	if err := NewMessagePublish(conn); err != nil {
		panic(err)
	}
}

func NewMessagePublish(conn *rabbitmq.Connection) error {
	p := person{Name: "rs", Age: 22}
	q := person{Name: "reza", Age: 23}
	if err := conn.Publish("exchange1", "rk", rabbitmq.PublishingOptions{}, p); err != nil {
		return err
	}
	if err := conn.Publish("exchange1", "rk2", rabbitmq.PublishingOptions{}, q); err != nil {
		return err
	}
	return nil
}
