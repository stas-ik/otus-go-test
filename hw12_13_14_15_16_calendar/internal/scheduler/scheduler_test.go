package scheduler

import (
	"context"
	"testing"
	"time"

	"github.com/stas-ik/otus-go-test/hw12_13_14_15_16_calendar/internal/logger"
	"github.com/stas-ik/otus-go-test/hw12_13_14_15_16_calendar/internal/rabbitmq"
	"github.com/stas-ik/otus-go-test/hw12_13_14_15_16_calendar/internal/storage"
)

type MockStorage struct {
	eventsToNotify  []storage.Event
	notifiedIDs     []string
	deleteOldCalled bool
}

func (m *MockStorage) GetEventsToNotify(ctx context.Context) ([]storage.Event, error) {
	return m.eventsToNotify, nil
}

func (m *MockStorage) MarkEventNotified(ctx context.Context, id string) error {
	m.notifiedIDs = append(m.notifiedIDs, id)
	return nil
}

func (m *MockStorage) DeleteOldEvents(ctx context.Context, olderThan time.Time) error {
	m.deleteOldCalled = true
	return nil
}

type MockPublisher struct {
	published []rabbitmq.Notification
}

func (m *MockPublisher) Publish(n rabbitmq.Notification) error {
	m.published = append(m.published, n)
	return nil
}

func TestScheduler_ProcessNotifications(t *testing.T) {
	ms := &MockStorage{
		eventsToNotify: []storage.Event{
			{ID: "1", Title: "Event 1", UserID: "user1", StartTime: time.Now()},
		},
	}
	mp := &MockPublisher{}
	log := logger.New("ERROR")
	s := New(ms, mp, log)

	s.ProcessNotifications(context.Background())

	if len(mp.published) != 1 {
		t.Errorf("expected 1 published notification, got %d", len(mp.published))
	}
	if mp.published[0].EventID != "1" {
		t.Errorf("expected EventID 1, got %s", mp.published[0].EventID)
	}
	if len(ms.notifiedIDs) != 1 || ms.notifiedIDs[0] != "1" {
		t.Errorf("expected event 1 to be marked as notified")
	}
}

func TestScheduler_ProcessCleanup(t *testing.T) {
	ms := &MockStorage{}
	mp := &MockPublisher{}
	log := logger.New("ERROR")
	s := New(ms, mp, log)

	s.ProcessCleanup(context.Background())

	if !ms.deleteOldCalled {
		t.Errorf("expected DeleteOldEvents to be called")
	}
}
