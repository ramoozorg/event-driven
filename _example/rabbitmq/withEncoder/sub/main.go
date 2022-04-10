package main

import (
	"fmt"
	"git.ramooz.org/ramooz/golang-components/event-driven/rabbitmq"
	"github.com/streadway/amqp"
	"go.mongodb.org/mongo-driver/bson"
	"log"
	"time"
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
	delivries, err := conn.Consume()
	if err != nil {
		panic(err)
	}

	enc, _ := rabbitmq.NewEncodedConn(conn, rabbitmq.BSON_ENCODER)
	p := Person{}
	for {
		for k, d := range delivries {
			enc.HandleConsumedDeliveries(k, p, d, messageHandler)
		}
		time.Sleep(10 * time.Second)
	}
}

func messageHandler(e *rabbitmq.EncodedConn, deliveries <-chan amqp.Delivery) {
	for d := range deliveries {
		log.Printf("Got message from queue %v and routing key %v\n", e.Conn.ConnOpt.Queues, d.RoutingKey)
		p := Person{}
		_ = bson.Unmarshal(d.Body, &p)
		fmt.Printf("%v - %T\n", p, p)
	}
}
