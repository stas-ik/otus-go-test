package memorystorage

import (
	"context"
	"sync"
	"time"

	"github.com/stas-ik/otus-go-test/hw12_13_14_15_16_calendar/internal/storage"
)

type Storage struct {
	mu     sync.RWMutex
	events map[string]storage.Event
}

func New() *Storage {
	return &Storage{
		events: make(map[string]storage.Event),
	}
}

func (s *Storage) CreateEvent(_ context.Context, event storage.Event) error {
	if event.ID == "" || event.Title == "" {
		return storage.ErrInvalidEvent
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.isTimeBusyLocked(event.ID, event.UserID, event.StartTime, event.EndTime) {
		return storage.ErrDateBusy
	}

	s.events[event.ID] = event
	return nil
}

func (s *Storage) UpdateEvent(_ context.Context, id string, event storage.Event) error {
	if event.Title == "" {
		return storage.ErrInvalidEvent
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.events[id]; !exists {
		return storage.ErrEventNotFound
	}

	if s.isTimeBusyLocked(id, event.UserID, event.StartTime, event.EndTime) {
		return storage.ErrDateBusy
	}

	event.ID = id
	s.events[id] = event
	return nil
}

func (s *Storage) DeleteEvent(_ context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.events[id]; !exists {
		return storage.ErrEventNotFound
	}

	delete(s.events, id)
	return nil
}

func (s *Storage) GetEventByID(_ context.Context, id string) (*storage.Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	event, exists := s.events[id]
	if !exists {
		return nil, storage.ErrEventNotFound
	}

	return &event, nil
}

func (s *Storage) ListEventsForDay(_ context.Context, date time.Time) ([]storage.Event, error) {
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	return s.listEventsBetween(startOfDay, endOfDay), nil
}

func (s *Storage) ListEventsForWeek(_ context.Context, startDate time.Time) ([]storage.Event, error) {
	startOfWeek := time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, startDate.Location())
	endOfWeek := startOfWeek.Add(7 * 24 * time.Hour)

	return s.listEventsBetween(startOfWeek, endOfWeek), nil
}

func (s *Storage) ListEventsForMonth(_ context.Context, startDate time.Time) ([]storage.Event, error) {
	startOfMonth := time.Date(startDate.Year(), startDate.Month(), 1, 0, 0, 0, 0, startDate.Location())
	endOfMonth := startOfMonth.AddDate(0, 1, 0)

	return s.listEventsBetween(startOfMonth, endOfMonth), nil
}

func (s *Storage) listEventsBetween(start, end time.Time) []storage.Event {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []storage.Event
	for _, event := range s.events {
		if (event.StartTime.Equal(start) || event.StartTime.After(start)) &&
			event.StartTime.Before(end) {
			result = append(result, event)
		}
	}

	return result
}

func (s *Storage) GetEventsToNotify(_ context.Context) ([]storage.Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []storage.Event
	now := time.Now()
	for _, event := range s.events {
		if !event.Notified && event.NotifyAt != nil && (event.NotifyAt.Before(now) || event.NotifyAt.Equal(now)) {
			result = append(result, event)
		}
	}
	return result, nil
}

func (s *Storage) MarkEventNotified(_ context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	event, exists := s.events[id]
	if !exists {
		return storage.ErrEventNotFound
	}
	event.Notified = true
	s.events[id] = event
	return nil
}

func (s *Storage) DeleteOldEvents(_ context.Context, olderThan time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for id, event := range s.events {
		if event.StartTime.Before(olderThan) {
			delete(s.events, id)
		}
	}
	return nil
}

func (s *Storage) isTimeBusyLocked(excludeID, userID string, start, end time.Time) bool {
	for _, event := range s.events {
		if event.ID == excludeID {
			continue
		}
		if event.UserID != userID {
			continue
		}

		if start.Before(event.EndTime) && end.After(event.StartTime) {
			return true
		}
	}
	return false
}
