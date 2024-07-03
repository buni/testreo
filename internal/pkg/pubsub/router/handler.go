//nolint:ireturn
package router

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/buni/wallet/internal/pkg/pubsub"
)

// Handler the signature of the handler function, where the business logic is implemented
// an example of how this should look can be found in the examples folder or the examples_test.go file.
type Handler[Event any] interface {
	HandlerNameFunc
	TopicFunc
	SubscriberOptions() []pubsub.SubscriberOption
	Handle(context.Context, *Event, pubsub.SubscriberMessage) error
}

type HandlerNameFunc interface {
	HandlerName() string
}

type TopicFunc interface {
	Topic() string
}

// HandleFunc is a wrapper around the Handler[Event any] interface.
type HandleFunc interface {
	Handle(ctx context.Context, msg pubsub.SubscriberMessage) error
}

// SubscribeFunc is a wrapper around the pubsub.Subscriber interface,
// that allows the handler to subscribe to a topic/queue, based on the SubscriberOptions provided by the handler implementation.
type SubscribeFunc interface {
	Subscribe(ctx context.Context) (<-chan pubsub.SubscriberMessage, error)
}

// HandleSubscribe is a union between HandleFunc and SubscribeFunc.
// this is what the router expects to receive when registering a handler.
type HandleSubscribe interface {
	HandlerNameFunc
	TopicFunc
	HandleFunc
	SubscribeFunc
}

// HandlerStruct is a generic struct that implements the HandleSubscribe interface.
// it takes a generic Handler implementation, and wraps it so it confronts to the HandleSubscribe interface.
type HandlerStruct[Event any] struct {
	Handler[Event]
	subscriber pubsub.Subscriber
	unmarshal  func(data []byte, v any) error
}

// NewJSONHandler is a helper function that creates a new HandlerStruct[Event] with the json.Unmarshal function as the unmarshaler.
func NewJSONHandler[Event any](handler Handler[Event], subscriber pubsub.Subscriber, middlewares ...Middleware) HandleSubscribe {
	return NewHandler(handler, subscriber, json.Unmarshal, middlewares...)
}

// NewHandler creates a new generic HandlerStruct[Event] instance, that implements the HandleSubscribe interface.
// it takes a generic Handler implementation, a pubsub.Subscriber implementation, a unmarshaler function, and a list of middleware.
// the unmarshaler function is used to unmarshal the message payload into the event struct, that is provided to the Handle[Event any] function.
// middleware are applied left to right, so the first middleware in the list is the first to be applied.
func NewHandler[Event any](
	handler Handler[Event],
	subscriber pubsub.Subscriber,
	unmarshaler func(data []byte, v any) error,
	middlewares ...Middleware,
) HandleSubscribe {
	hs := &HandlerStruct[Event]{
		Handler:    handler,
		subscriber: subscriber,
		unmarshal:  unmarshaler,
	}

	handleSubscribeUnion := applyMiddleware(hs, middlewares...)

	return handleSubscribeUnion
}

// Handle is the implementation of the HandleFunc interface.
// it unmarshals the message payload into the Event struct, and calls the Handle(context.Context, *Event, pubsub.SubscriptionMessage) error function.
func (hs *HandlerStruct[Event]) Handle(ctx context.Context, msg pubsub.SubscriberMessage) (err error) {
	event := new(Event)

	message, err := msg.Message(ctx)
	if err != nil {
		return fmt.Errorf("failed to get message: %w", err)
	}

	err = hs.unmarshal(message.Payload, event)
	if err != nil {
		return fmt.Errorf("failed to : %w", err)
	}

	err = hs.Handler.Handle(ctx, event, msg)
	if err != nil {
		return fmt.Errorf("failed to handle event: %w", err)
	}

	return nil
}

// Subscribe is the implementation of the SubscribeFunc interface.
// it calls the Subscribe(ctx context.Context, topic string, opts ...SubscriberOption) (<-chan SubscriptionMessage, error) function on the provided Subscriber
// with the topic and options provided by the Handler implementation.
func (hs *HandlerStruct[Event]) Subscribe(ctx context.Context) (<-chan pubsub.SubscriberMessage, error) {
	return hs.subscriber.Subscribe(ctx, hs.Handler.Topic(), hs.Handler.SubscriberOptions()...) //nolint:wrapcheck
}
