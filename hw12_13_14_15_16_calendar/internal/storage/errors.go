package storage

import "errors"

var (
	ErrEventNotFound = errors.New("event not found")

	ErrDateBusy = errors.New("date is busy")

	ErrInvalidEvent = errors.New("invalid event")
)
