package memorystorage

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stas-ik/otus-go-test/hw12_13_14_15_calendar/internal/storage" //nolint:depguard
)

func TestStorage_CreateEvent(t *testing.T) {
	s := New()
	ctx := context.Background()

	event := storage.Event{
		ID:        "1",
		Title:     "Test Event",
		StartTime: time.Now(),
		EndTime:   time.Now().Add(time.Hour),
		UserID:    "user1",
	}

	err := s.CreateEvent(ctx, event)
	if err != nil {
		t.Fatalf("CreateEvent failed: %v", err)
	}

	// проверяем, что событие создано
	retrieved, err := s.GetEventByID(ctx, "1")
	if err != nil {
		t.Fatalf("GetEventByID failed: %v", err)
	}

	if retrieved.Title != event.Title {
		t.Errorf("Expected title %s, got %s", event.Title, retrieved.Title)
	}
}

func TestStorage_CreateEvent_InvalidEvent(t *testing.T) {
	s := New()
	ctx := context.Background()

	event := storage.Event{
		ID:    "",
		Title: "Test Event",
	}

	err := s.CreateEvent(ctx, event)
	if !errors.Is(err, storage.ErrInvalidEvent) {
		t.Errorf("Expected ErrInvalidEvent, got %v", err)
	}
}

func TestStorage_CreateEvent_DateBusy(t *testing.T) {
	s := New()
	ctx := context.Background()

	start := time.Now()
	end := start.Add(time.Hour)

	event1 := storage.Event{
		ID:        "1",
		Title:     "Event 1",
		StartTime: start,
		EndTime:   end,
		UserID:    "user1",
	}

	err := s.CreateEvent(ctx, event1)
	if err != nil {
		t.Fatalf("CreateEvent failed: %v", err)
	}

	// пытаемся создать пересекающееся событие
	event2 := storage.Event{
		ID:        "2",
		Title:     "Event 2",
		StartTime: start.Add(30 * time.Minute),
		EndTime:   end.Add(30 * time.Minute),
		UserID:    "user1",
	}

	err = s.CreateEvent(ctx, event2)
	if !errors.Is(err, storage.ErrDateBusy) {
		t.Errorf("Expected ErrDateBusy, got %v", err)
	}

	// событие для другого пользователя должно создаться
	event3 := storage.Event{
		ID:        "3",
		Title:     "Event 3",
		StartTime: start.Add(30 * time.Minute),
		EndTime:   end.Add(30 * time.Minute),
		UserID:    "user2",
	}

	err = s.CreateEvent(ctx, event3)
	if err != nil {
		t.Errorf("CreateEvent for different user should succeed: %v", err)
	}
}

func TestStorage_UpdateEvent(t *testing.T) {
	s := New()
	ctx := context.Background()

	event := storage.Event{
		ID:        "1",
		Title:     "Original Title",
		StartTime: time.Now(),
		EndTime:   time.Now().Add(time.Hour),
		UserID:    "user1",
	}

	_ = s.CreateEvent(ctx, event)

	// обновляем событие
	updatedEvent := event
	updatedEvent.Title = "Updated Title"

	err := s.UpdateEvent(ctx, "1", updatedEvent)
	if err != nil {
		t.Fatalf("UpdateEvent failed: %v", err)
	}

	retrieved, _ := s.GetEventByID(ctx, "1")
	if retrieved.Title != "Updated Title" {
		t.Errorf("Expected title 'Updated Title', got %s", retrieved.Title)
	}
}

func TestStorage_UpdateEvent_NotFound(t *testing.T) {
	s := New()
	ctx := context.Background()

	event := storage.Event{
		ID:        "999",
		Title:     "Non-existent",
		StartTime: time.Now(),
		EndTime:   time.Now().Add(time.Hour),
		UserID:    "user1",
	}

	err := s.UpdateEvent(ctx, "999", event)
	if !errors.Is(err, storage.ErrEventNotFound) {
		t.Errorf("Expected ErrEventNotFound, got %v", err)
	}
}

func TestStorage_DeleteEvent(t *testing.T) {
	s := New()
	ctx := context.Background()

	event := storage.Event{
		ID:        "1",
		Title:     "Test Event",
		StartTime: time.Now(),
		EndTime:   time.Now().Add(time.Hour),
		UserID:    "user1",
	}

	_ = s.CreateEvent(ctx, event)

	err := s.DeleteEvent(ctx, "1")
	if err != nil {
		t.Fatalf("DeleteEvent failed: %v", err)
	}

	_, err = s.GetEventByID(ctx, "1")
	if !errors.Is(err, storage.ErrEventNotFound) {
		t.Errorf("Expected ErrEventNotFound after deletion, got %v", err)
	}
}

func TestStorage_ListEventsForDay(t *testing.T) {
	s := New()
	ctx := context.Background()

	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 10, 0, 0, 0, now.Location())
	tomorrow := today.Add(24 * time.Hour)

	event1 := storage.Event{
		ID:        "1",
		Title:     "Today's Event",
		StartTime: today,
		EndTime:   today.Add(time.Hour),
		UserID:    "user1",
	}

	event2 := storage.Event{
		ID:        "2",
		Title:     "Tomorrow's Event",
		StartTime: tomorrow,
		EndTime:   tomorrow.Add(time.Hour),
		UserID:    "user1",
	}

	_ = s.CreateEvent(ctx, event1)
	_ = s.CreateEvent(ctx, event2)

	events, err := s.ListEventsForDay(ctx, today)
	if err != nil {
		t.Fatalf("ListEventsForDay failed: %v", err)
	}

	if len(events) != 1 {
		t.Errorf("Expected 1 event for today, got %d", len(events))
	}

	if len(events) > 0 && events[0].Title != "Today's Event" {
		t.Errorf("Expected 'Today's Event', got %s", events[0].Title)
	}
}

func TestStorage_ConcurrentAccess(t *testing.T) {
	t.Parallel()
	s := New()
	ctx := context.Background()

	var wg sync.WaitGroup
	iterations := 100

	// параллельное создание событий
	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			event := storage.Event{
				ID:        string(rune('A' + id)),
				Title:     "Concurrent Event",
				StartTime: time.Now().Add(time.Duration(id) * time.Hour),
				EndTime:   time.Now().Add(time.Duration(id+1) * time.Hour),
				UserID:    "user1",
			}

			_ = s.CreateEvent(ctx, event)
		}(i)
	}

	wg.Wait()

	// параллельное чтение
	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			_, _ = s.GetEventByID(ctx, string(rune('A'+id)))
		}(i)
	}

	wg.Wait()
}
