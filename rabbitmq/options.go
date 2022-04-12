package rabbitmq

import "fmt"

// Options for new connection of rabbitmq
type Options struct {
	UriAddress      string   // UriAddress of rabbitmq, amqp://user:password@x.x.x.x:port
	Exchange        string   // Exchange name
	RoutingKeys     []string // RoutingKeys for messages
	Queues          []string // Queues name
	Kind            string
	DurableExchange bool
	AutoAck         bool
	AutoDelete      bool
	NoWait          bool
	ExclusiveQueue  bool
}

// getDefaultOptions create default options
func getDefaultOptions() *Options {
	return &Options{
		Exchange:        "NewData",
		Queues:          []string{"defaultQueue"},
		Kind:            "topic",
		DurableExchange: true,
		AutoAck:         true,
		AutoDelete:      false,
		NoWait:          false,
		ExclusiveQueue:  false,
	}
}

// CreateURIAddress create url address from input configuration
func CreateURIAddress(username, password, address, vhost string) string {
	return fmt.Sprintf("amqp://%s:%s@%s/%s", username, password, address, vhost)
}
