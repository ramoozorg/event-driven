package rabbitmq

import "errors"

var (
	SERVICE_NAME_ERROR            = errors.New("service name is empty")
	URI_ADDRESS_ERROR             = errors.New("uri address is invalid, please enter amqp://guest:guest@localhost:5672 for example")
	ROUTING_KEYS_EMPTY_ERROR      = errors.New("routing keys is empty")
	CONNECTION_CLOSED_ERROR       = errors.New("rabbitMQ connection closed, try to reconnect")
	NIL_CCONECTION_ERROR          = errors.New("nil rabbitmq connection")
	EXCHANGE_ALREADY_EXISTS_ERROR = errors.New("exchange already declared")
	QUEUE_ALREADY_EXISTS_ERROR    = errors.New("queue already declared")
	EXHCNAGE_NOT_FOUND_ERROR      = errors.New("exchange not declare")
)
