package storage

import (
	"context"
	"time"
)

type Storage interface {
	CreateEvent(ctx context.Context, event Event) error
	UpdateEvent(ctx context.Context, id string, event Event) error
	DeleteEvent(ctx context.Context, id string) error
	GetEventByID(ctx context.Context, id string) (*Event, error)
	ListEventsForDay(ctx context.Context, date time.Time) ([]Event, error)
	ListEventsForWeek(ctx context.Context, startDate time.Time) ([]Event, error)
	ListEventsForMonth(ctx context.Context, startDate time.Time) ([]Event, error)
	GetEventsToNotify(ctx context.Context) ([]Event, error)
	MarkEventNotified(ctx context.Context, id string) error
	DeleteOldEvents(ctx context.Context, olderThan time.Time) error
}
