package pubsub

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
)

var ErrFailedToScanMessage = errors.New("failed to scan message")

//go:generate mockgen -source=publisher.go -destination=mock/publisher_mocks.go -package mock_pubsub

// Publisher is the interface that every publisher should implement.
type Publisher interface {
	Publish(ctx context.Context, msg *Message, opts ...PublishOption) error
}

type PublisherWithEvent interface {
	Publisher
}

type Headers map[string]string

// Message holds the payload and headers of an event.
type Message struct {
	ID      *string
	Key     string
	Topic   string
	Payload []byte
	Headers Headers
}

// NewJSONMessage creates a new Message with the payload marshaled to JSON.
func NewJSONMessage(payload any, headers Headers) (*Message, error) {
	b, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	if headers == nil {
		headers = Headers{}
	}

	return &Message{
		ID:      nil,
		Key:     "",
		Topic:   "",
		Payload: b,
		Headers: headers,
	}, nil
}

// Value ...
func (a Message) Value() (driver.Value, error) {
	return json.Marshal(a) //nolint: wrapcheck,musttag
}

// Scan ...
func (a *Message) Scan(value any) error {
	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, a) //nolint: wrapcheck,musttag
	case string:
		return json.Unmarshal([]byte(v), a) //nolint: wrapcheck,musttag
	default:
		return ErrFailedToScanMessage
	}
}
