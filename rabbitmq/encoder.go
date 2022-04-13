package rabbitmq

import (
	"git.ramooz.org/ramooz/golang-components/event-driven/rabbitmq/enc"
	"sync"
)

var encMap map[string]Encoder
var encLock sync.Mutex

// Encoder interface is for all register encoders
type Encoder interface {
	Encode(msg interface{}) ([]byte, error)
	Decode(data []byte, vPtr interface{}) error
}

const (
	JSON_ENCODER = "json"
	BSON_ENCODER = "bson"
)

func init() {
	encMap = make(map[string]Encoder)
	// Register json, bson encoder
	RegisterEncoder(JSON_ENCODER, &enc.JsonEncoder{})
	RegisterEncoder(BSON_ENCODER, &enc.BsonEncoder{})
}

type EncodedConn struct {
	Conn *Connection
	Enc  Encoder
}

// RegisterEncoder will register the encType with the given Encoder. Useful for customization.
func RegisterEncoder(encType string, enc Encoder) {
	encLock.Lock()
	defer encLock.Unlock()
	encMap[encType] = enc
}

// EncoderForType will return the registered Encoder for the encType.
func EncoderForType(encType string) Encoder {
	encLock.Lock()
	defer encLock.Unlock()
	return encMap[encType]
}
