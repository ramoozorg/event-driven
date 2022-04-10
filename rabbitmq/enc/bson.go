package enc

import (
	"go.mongodb.org/mongo-driver/bson"
	"strings"
)

// BsonEncoder is a BSON Encoder implementation for EncodedConn.
// This encoder will use the mongo bson encoder to Marshal
// and Unmarshal most types, including structs.
type BsonEncoder struct{}

// Encode
func (be *BsonEncoder) Encode(msg interface{}) ([]byte, error) {
	b, err := bson.Marshal(msg)
	if err != nil {
		return nil, err
	}
	return b, nil
}

// Decode
func (be *BsonEncoder) Decode(data []byte, vPtr interface{}) (err error) {
	switch arg := vPtr.(type) {
	case *string:
		str := string(data)
		if strings.HasPrefix(str, `"`) && strings.HasSuffix(str, `"`) {
			*arg = str[1 : len(str)-1]
		} else {
			*arg = str
		}
	case *[]byte:
		*arg = data
	default:
		err = bson.Unmarshal(data, arg)
	}
	return
}
