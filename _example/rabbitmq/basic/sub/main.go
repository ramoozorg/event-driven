package main

import (
	"encoding/json"
	"fmt"
	"git.ramooz.org/ramooz/golang-components/event-driven/rabbitmq"
	"github.com/streadway/amqp"
	"log"
	"time"
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
	delivries, err := conn.Consume()
	if err != nil {
		panic(err)
	}
	for {
		for k, d := range delivries {
			conn.HandleConsumedDeliveries(k, d, messageHandler)
		}
		time.Sleep(10 * time.Second)
	}
}

func messageHandler(c *rabbitmq.Connection, deliveries <-chan amqp.Delivery) {
	for d := range deliveries {
		//handle the custom message
		log.Printf("Got message from queue %v and routing key %v\n", c.ConnOpt.Queues, d.RoutingKey)
		var v interface{}
		_ = json.Unmarshal(d.Body, &v)
		fmt.Printf("%v\n", v)
	}
}
