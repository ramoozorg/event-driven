package main

import (
	"fmt"
	"git.ramooz.org/ramooz/golang-components/event-driven/rabbitmq"
	"go.mongodb.org/mongo-driver/bson"
	"time"
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

	for {
		if conn.IsConnected() {
			break
		}
		time.Sleep(1 * time.Second)
	}

	if err := conn.ExchangeDeclare("exchange1", rabbitmq.TOPIC); err != nil {
		panic(err)
	}
	if err := conn.QueueDeclare("queue1", "exchange1", "rk", messageHandler); err != nil {
		panic(err)
	}

	go func() {
		for {
			if err := conn.Consume(); err != nil {
				panic(err)
			}
			time.Sleep(5 * time.Second)
		}
	}()
	<-done
}

func messageHandler(delivery rabbitmq.Delivery) error {
	msg := <-delivery
	p := person{}
	_ = bson.Unmarshal(msg.Body, &p)
	fmt.Printf("New Message from exchange %v routingKey %v with body %v received\n", msg.Exchange, msg.RoutingKey, p)
	return nil
}
