package app

import (
	"context"
	"time"

	"github.com/stas-ik/otus-go-test/hw12_13_14_15_16_calendar/internal/storage" //nolint:depguard
)

type App struct {
	logger  Logger
	storage Storage
}

type Logger interface {
	Debug(msg string)
	Info(msg string)
	Warn(msg string)
	Error(msg string)
	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
}

type Storage interface {
	CreateEvent(ctx context.Context, event storage.Event) error
	UpdateEvent(ctx context.Context, id string, event storage.Event) error
	DeleteEvent(ctx context.Context, id string) error
	GetEventByID(ctx context.Context, id string) (*storage.Event, error)
	ListEventsForDay(ctx context.Context, date time.Time) ([]storage.Event, error)
	ListEventsForWeek(ctx context.Context, startDate time.Time) ([]storage.Event, error)
	ListEventsForMonth(ctx context.Context, startDate time.Time) ([]storage.Event, error)
}

func New(logger Logger, storage Storage) *App {
	return &App{
		logger:  logger,
		storage: storage,
	}
}

func (a *App) CreateEvent(ctx context.Context, event storage.Event) error {
	a.logger.Debugf("Creating event: %s", event.ID)
	return a.storage.CreateEvent(ctx, event)
}

func (a *App) UpdateEvent(ctx context.Context, id string, event storage.Event) error {
	a.logger.Debugf("Updating event: %s", id)
	return a.storage.UpdateEvent(ctx, id, event)
}

func (a *App) DeleteEvent(ctx context.Context, id string) error {
	a.logger.Debugf("Deleting event: %s", id)
	return a.storage.DeleteEvent(ctx, id)
}

func (a *App) GetEventByID(ctx context.Context, id string) (*storage.Event, error) {
	a.logger.Debugf("Getting event: %s", id)
	return a.storage.GetEventByID(ctx, id)
}

func (a *App) ListEventsForDay(ctx context.Context, date time.Time) ([]storage.Event, error) {
	a.logger.Debugf("Listing events for day: %s", date.Format("2006-01-02"))
	return a.storage.ListEventsForDay(ctx, date)
}

func (a *App) ListEventsForWeek(ctx context.Context, startDate time.Time) ([]storage.Event, error) {
	a.logger.Debugf("Listing events for week starting: %s", startDate.Format("2006-01-02"))
	return a.storage.ListEventsForWeek(ctx, startDate)
}

func (a *App) ListEventsForMonth(ctx context.Context, startDate time.Time) ([]storage.Event, error) {
	a.logger.Debugf("Listing events for month starting: %s", startDate.Format("2006-01-02"))
	return a.storage.ListEventsForMonth(ctx, startDate)
}
