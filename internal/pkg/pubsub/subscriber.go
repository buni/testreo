package pubsub

import (
	"context"
)

type Subscriber interface {
	Subscribe(ctx context.Context, topic string, opts ...SubscriberOption) (<-chan SubscriberMessage, error)
}

type SubscriberMessage interface {
	Message(context.Context) (*Message, error)
	IsProcessed(context.Context) (bool, error)
	Ack(context.Context) error
	Nack(context.Context) error
	DLQ(context.Context) error
}
