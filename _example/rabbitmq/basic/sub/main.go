package main

import (
	"fmt"
	"git.ramooz.org/ramooz/golang-components/event-driven/rabbitmq"
	"go.mongodb.org/mongo-driver/bson"
	"os"
	"time"
)

var (
	routingKeys = []string{"rk1", "rk2"}
	queues      = []string{"queue1", "queue2"}
)

type person struct {
	Name string `bson:"name" json:"name"`
	Age  int    `bson:"age" json:"age"`
}

func main() {
	chSignal := make(chan os.Signal, 1)
	done := make(chan bool, 1)
	conn, err := rabbitmq.NewConnection("test", &rabbitmq.Options{
		UriAddress:      "amqp://guest:guest@localhost:5672",
		DurableExchange: true,
		AutoAck:         true,
		ExclusiveQueue:  false,
	}, chSignal)
	if err != nil {
		panic(err)
	}

	for {
		if conn.IsConnected() {
			break
		}
		time.Sleep(1 * time.Second)
	}

	if err := conn.ExchangeDeclare(map[string]string{"exch1": "topic", "exch2": "topic"}); err != nil {
		panic(err)
	}
	if err := conn.QueueDeclare(queues); err != nil {
		panic(err)
	}
	if err := conn.RoutingKeyDeclare(routingKeys); err != nil {
		panic(err)
	}
	if err := conn.QueueBind(); err != nil {
		panic(err)
	}

	go func() {
		for {
			delivries, err := conn.Consume()
			if err != nil {
				panic(err)
			}
			for queue, msg := range delivries {
				msg := conn.MessageHandler(queue, msg)
				if !msg.IsEmpty() {
					p := person{}
					_ = bson.Unmarshal(msg.Body, &p)
					fmt.Printf("New Message from exchange %v queue %v routingKey %v with body %v received\n", msg.Exchange, msg.Queue, msg.RoutingKey, p)
				}
			}
		}
	}()

	go func() {
		sig := <-chSignal
		fmt.Println()
		fmt.Println(sig)
		done <- true
	}()
	<-done
}
