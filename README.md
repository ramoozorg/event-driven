# Event-Driven

[![Go Reference](https://pkg.go.dev/badge/github.com/ramoozorg/event-driven.svg)](https://pkg.go.dev/github.com/ramoozorg/event-driven)

Event-driven architecture is a software architecture and model for application design. With an event-driven system, the capture, communication, processing, and persistence of events are the core structure of the solution. This differs from a traditional request-driven model.

## Features
- broadcast event via rabbitMQ
- auto reconnect pattern for rabbitMQ
- fault tolerance on panic

### Example Publish Message

```go
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
	if err := conn.Publish("exchange1", "rk2", q, rabbitmq.PublishingOptions{}); err != nil {
		return err
	}
	return nil
}

```

### Example Consume

```go
package main

import (
	"fmt"
	"git.ramooz.org/ramooz/golang-components/event-driven/rabbitmq"
	"go.mongodb.org/mongo-driver/bson"
)

type person struct {
	Name string `bson:"name" json:"name"`
	Age  int    `bson:"age" json:"age"`
}

func main() {
	done := make(chan bool, 1)
	conn, err := rabbitmq.NewConnection("test", &rabbitmq.Options{
		UriAddress:      rabbitmq.CreateURIAddress("guest", "guest", "localhost:5672", ""),
		DurableExchange: true,
		AutoAck:         true,
		ExclusiveQueue:  false,
	}, nil)
	if err != nil {
		panic(err)
	}

	if err := conn.ExchangeDeclare("exchange1", rabbitmq.TOPIC); err != nil {
		panic(err)
	}
	if err := conn.DeclareConsumerQueue("queue1", "exchange1", "rk", eventHandler); err != nil {
		panic(err)
	}
	if err := conn.DeclareConsumerQueue("queue2", "exchange1", "rk2", eventHandler); err != nil {
		panic(err)
	}

	if err := conn.Consume(); err != nil {
		panic(err)
	}
	<-done
}

func eventHandler(queue string, delivery rabbitmq.Delivery) {
	p := person{}
	_ = bson.Unmarshal(delivery.Body, &p)
	fmt.Printf("New Event from exchange %v queue %v routingKey %v with body %v received\n", delivery.Exchange, queue, delivery.RoutingKey, p)
}
```
