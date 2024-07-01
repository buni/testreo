package pubsub

import (
	"errors"
)

const (
	InvalidOptionAlias = iota
	PublisherOptionAlias
	PublishOptionAlias
	SubscriberOptionAlias

	PublishOptionsStructOptionType = -1
)

var (
	ErrInvalidOptionType        = errors.New("invalid option type")
	ErrInvalidPublishAtValue    = errors.New("invalid publish at value")
	ErrInvalidPublishAfterValue = errors.New("invalid publish after value")
	ErrInvalidRescheduleValue   = errors.New("invalid reschedule value")
)

// OptionType is the integer type of the option
// this is used to infer the option type.
type OptionType int

// Alias ...
type OptionAlias int

// Option defines the methods that an option must implement.
type Option interface {
	Type() OptionType
	Alias() OptionAlias
	Value() any
}

// PublisherOption is an alias for Option, that should only be passed to publishers.
type PublisherOption Option

// PublishOption is an alias for an Option that should only be passed to a Publisher.
type PublishOption Option

// SubscriberOption is an alias for Option, that should only be passed to subscribers.
type SubscriberOption Option

type OptionValue struct {
	OptionValue any
	OptionAlias OptionAlias
	OptionType  OptionType
}

func (o OptionValue) Type() OptionType {
	return o.OptionType
}

func (o OptionValue) Value() any {
	return o.OptionValue
}

func (o OptionValue) Alias() OptionAlias {
	return o.OptionAlias
}
