package app

import (
	"context"
	"testing"
	"time"

	"github.com/stas-ik/otus-go-test/hw12_13_14_15_16_calendar/internal/storage"
)

type mockStorage struct {
	events map[string]storage.Event
	err    error
}

func (m *mockStorage) CreateEvent(ctx context.Context, event storage.Event) error {
	if m.err != nil {
		return m.err
	}
	m.events[event.ID] = event
	return nil
}

func (m *mockStorage) UpdateEvent(ctx context.Context, id string, event storage.Event) error {
	if m.err != nil {
		return m.err
	}
	m.events[id] = event
	return nil
}

func (m *mockStorage) DeleteEvent(ctx context.Context, id string) error {
	if m.err != nil {
		return m.err
	}
	delete(m.events, id)
	return nil
}

func (m *mockStorage) GetEventByID(ctx context.Context, id string) (*storage.Event, error) {
	if m.err != nil {
		return nil, m.err
	}
	ev, ok := m.events[id]
	if !ok {
		return nil, storage.ErrEventNotFound
	}
	return &ev, nil
}

func (m *mockStorage) ListEventsForDay(ctx context.Context, date time.Time) ([]storage.Event, error) {
	return nil, m.err
}

func (m *mockStorage) ListEventsForWeek(ctx context.Context, startDate time.Time) ([]storage.Event, error) {
	return nil, m.err
}

func (m *mockStorage) ListEventsForMonth(ctx context.Context, startDate time.Time) ([]storage.Event, error) {
	return nil, m.err
}

func (m *mockStorage) GetEventsToNotify(ctx context.Context) ([]storage.Event, error) {
	return nil, m.err
}

func (m *mockStorage) MarkEventNotified(ctx context.Context, id string) error {
	return m.err
}

func (m *mockStorage) DeleteOldEvents(ctx context.Context, olderThan time.Time) error {
	return m.err
}

type mockLogger struct{}

func (m *mockLogger) Debug(msg string)                          {}
func (m *mockLogger) Info(msg string)                           {}
func (m *mockLogger) Warn(msg string)                           {}
func (m *mockLogger) Error(msg string)                          {}
func (m *mockLogger) Debugf(format string, args ...interface{}) {}
func (m *mockLogger) Infof(format string, args ...interface{})  {}
func (m *mockLogger) Warnf(format string, args ...interface{})  {}
func (m *mockLogger) Errorf(format string, args ...interface{}) {}

func TestApp(t *testing.T) {
	ms := &mockStorage{events: make(map[string]storage.Event)}
	ml := &mockLogger{}
	a := New(ml, ms)

	ctx := context.Background()
	event := storage.Event{ID: "1", Title: "Test"}

	t.Run("CreateEvent", func(t *testing.T) {
		err := a.CreateEvent(ctx, event)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if len(ms.events) != 1 {
			t.Errorf("expected 1 event, got %d", len(ms.events))
		}
	})

	t.Run("GetEventByID", func(t *testing.T) {
		ev, err := a.GetEventByID(ctx, "1")
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if ev.Title != "Test" {
			t.Errorf("expected title Test, got %s", ev.Title)
		}
	})

	t.Run("UpdateEvent", func(t *testing.T) {
		event.Title = "Updated"
		err := a.UpdateEvent(ctx, "1", event)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if ms.events["1"].Title != "Updated" {
			t.Errorf("expected title Updated, got %s", ms.events["1"].Title)
		}
	})

	t.Run("DeleteEvent", func(t *testing.T) {
		err := a.DeleteEvent(ctx, "1")
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if len(ms.events) != 0 {
			t.Errorf("expected 0 events, got %d", len(ms.events))
		}
	})
}
