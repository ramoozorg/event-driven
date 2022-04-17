package main

import (
	"fmt"

	"github.com/ramoozorg/event-driven/rabbitmq"
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
