package rabbitmq

// message is the amqp request to publish
type message struct {
	RoutingKey    string
	ReplyTo       string
	ContentType   string
	CorrelationID string
	Priority      uint8
	Body          []byte `bson:"body" json:"body"`
}

// Message is consumed message structure
type Message struct {
	Exchange   string
	Queue      string
	RoutingKey string
	Body       []byte `bson:"body" json:"body"`
}
