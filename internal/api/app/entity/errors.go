package entity

import "errors"

var (
	ErrUnsupportedEventVersion = errors.New("unsupported event version")
	ErrUnsupportedEventType    = errors.New("unsupported event type")
	ErrOutOfOrderWalletEvent   = errors.New("out of order wallet event")
	ErrInvalidEventType        = errors.New("invalid event type")
)
