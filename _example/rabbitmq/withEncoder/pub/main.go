package main

import (
	"git.ramooz.org/ramooz/golang-components/event-driven/rabbitmq"
)

type Person struct {
	Name string `json:"name" bson:"name"`
	Age  int    `json:"age" bson:"age"`
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
	enc, err := rabbitmq.NewEncodedConn(conn, rabbitmq.BSON_ENCODER)
	if err != nil {
		panic(err)
	}
	if err := NewMessageEncodedBsonPublish(enc); err != nil {
		panic(err)
	}
}

func NewMessageEncodedJsonPublish(conn *rabbitmq.EncodedConn) error {
	for _, key := range conn.Conn.ConnOpt.RoutingKeys {
		p := Person{Name: "javad", Age: 28}
		if err := conn.Publish(key, "", "", 0, p); err != nil {
			return err
		}
	}
	return nil
}

func NewMessageEncodedBsonPublish(conn *rabbitmq.EncodedConn) error {
	for _, key := range conn.Conn.ConnOpt.RoutingKeys {
		p := Person{Name: "ali", Age: 23}
		if err := conn.Publish(key, "", "", 0, p); err != nil {
			return err
		}
	}
	return nil
}
