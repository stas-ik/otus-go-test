package storage

import (
	"testing"
)

func TestErrors(t *testing.T) {
	if ErrEventNotFound.Error() != "event not found" {
		t.Errorf("unexpected error message: %s", ErrEventNotFound.Error())
	}
	if ErrDateBusy.Error() != "date is busy" {
		t.Errorf("unexpected error message: %s", ErrDateBusy.Error())
	}
	if ErrInvalidEvent.Error() != "invalid event" {
		t.Errorf("unexpected error message: %s", ErrInvalidEvent.Error())
	}
}
