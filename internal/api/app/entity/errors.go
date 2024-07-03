package entity

import "errors"

var (
	ErrUnsupportedEventVersion = errors.New("unsupported event version")
	ErrUnsupportedEventType    = errors.New("unsupported event type")
	ErrInvalidEventType        = errors.New("invalid event type")
	ErrEntityNotFound          = errors.New("entity not found")
	ErrNegativeAmount          = errors.New("negative amount")
	ErrInsufficientBalance     = errors.New("insufficient balance")
)
