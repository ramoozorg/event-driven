package main

import (
	"git.ramooz.org/ramooz/golang-components/event-driven/rabbitmq"
	"log"
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

	if err := NewMessagePublish(conn); err != nil {
		panic(err)
	}
}

func NewMessagePublish(conn *rabbitmq.Connection) error {
	p := person{Name: "javad", Age: 28}
	if err := conn.Publish("rk1", "", "text/plain", "", 0, p); err != nil {
		return err
	}
	log.Println("message published ", p)
	return nil
}
