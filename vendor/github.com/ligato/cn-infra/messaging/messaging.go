package messaging

import (
	"github.com/golang/protobuf/proto"
	"github.com/ligato/cn-infra/db/keyval"
)

// BytesPublisher allows to publish a message of type []bytes into messaging system.
type BytesPublisher interface {
	Publish(key string, data []byte) error
}

// ProtoPublisher allows to publish a message of type proto.Message into messaging system.
type ProtoPublisher interface {
	Publish(key string, data proto.Message) error
}

// BytesMessage defines functions for inspection of a message received from messaging system.
type BytesMessage interface {
	keyval.BytesKvPair
}

// ProtoMessage defines functions for inspection of a message receive from messaging system.
type ProtoMessage interface {
	keyval.ProtoKvPair
}
