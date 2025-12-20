package internalhttp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stas-ik/otus-go-test/hw12_13_14_15_16_calendar/internal/storage"
)

type mockApplication struct {
	events map[string]storage.Event
	err    error
}

func (m *mockApplication) CreateEvent(_ context.Context, event storage.Event) error {
	if m.err != nil {
		return m.err
	}
	m.events[event.ID] = event
	return nil
}

func (m *mockApplication) UpdateEvent(_ context.Context, id string, event storage.Event) error {
	if m.err != nil {
		return m.err
	}
	if _, ok := m.events[id]; !ok {
		return storage.ErrEventNotFound
	}
	m.events[id] = event
	return nil
}

func (m *mockApplication) DeleteEvent(_ context.Context, id string) error {
	if m.err != nil {
		return m.err
	}
	delete(m.events, id)
	return nil
}

func (m *mockApplication) GetEventByID(_ context.Context, id string) (*storage.Event, error) {
	if m.err != nil {
		return nil, m.err
	}
	ev, ok := m.events[id]
	if !ok {
		return nil, storage.ErrEventNotFound
	}
	return &ev, nil
}

func (m *mockApplication) ListEventsForDay(_ context.Context, date time.Time) ([]storage.Event, error) {
	if m.err != nil {
		return nil, m.err
	}
	var res []storage.Event
	for _, e := range m.events {
		if e.StartTime.Format("2006-01-02") == date.Format("2006-01-02") {
			res = append(res, e)
		}
	}
	return res, nil
}

func (m *mockApplication) ListEventsForWeek(_ context.Context, _ time.Time) ([]storage.Event, error) {
	if m.err != nil {
		return nil, m.err
	}
	return nil, nil // Simple mock
}

func (m *mockApplication) ListEventsForMonth(_ context.Context, _ time.Time) ([]storage.Event, error) {
	if m.err != nil {
		return nil, m.err
	}
	return nil, nil // Simple mock
}

type mockLogger struct{}

func (m *mockLogger) Info(_ string)  {}
func (m *mockLogger) Error(_ string) {}

func TestServer_CreateEvent(t *testing.T) {
	mockApp := &mockApplication{events: make(map[string]storage.Event)}
	server := NewServer(&mockLogger{}, mockApp, "localhost", "8080")

	event := eventDTO{
		ID:        "1",
		Title:     "Test Event",
		StartTime: time.Now(),
		EndTime:   time.Now().Add(time.Hour),
		UserID:    "user1",
	}
	body, _ := json.Marshal(event)

	req := httptest.NewRequest(http.MethodPost, "/api/events", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	server.mux.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d", w.Code)
	}

	if len(mockApp.events) != 1 {
		t.Errorf("expected 1 event in mock storage, got %d", len(mockApp.events))
	}
}

func TestServer_GetEventByID(t *testing.T) {
	id := "123"
	event := storage.Event{ID: id, Title: "Existing Event"}
	mockApp := &mockApplication{events: map[string]storage.Event{id: event}}
	server := NewServer(&mockLogger{}, mockApp, "localhost", "8080")

	req := httptest.NewRequest(http.MethodGet, "/api/events/"+id, nil)
	w := httptest.NewRecorder()

	server.mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp eventDTO
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}
	if resp.Title != event.Title {
		t.Errorf("expected title %s, got %s", event.Title, resp.Title)
	}
}

func TestServer_DeleteEvent(t *testing.T) {
	id := "123"
	event := storage.Event{ID: id, Title: "To be deleted"}
	mockApp := &mockApplication{events: map[string]storage.Event{id: event}}
	server := NewServer(&mockLogger{}, mockApp, "localhost", "8080")

	req := httptest.NewRequest(http.MethodDelete, "/api/events/"+id, nil)
	w := httptest.NewRecorder()

	server.mux.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected status 204, got %d", w.Code)
	}

	if len(mockApp.events) != 0 {
		t.Errorf("expected 0 events, got %d", len(mockApp.events))
	}
}

func TestServer_ListEventsForDay(t *testing.T) {
	dateStr := "2023-10-27"
	date, _ := time.Parse("2006-01-02", dateStr)
	event := storage.Event{ID: "1", Title: "Event 1", StartTime: date}
	mockApp := &mockApplication{events: map[string]storage.Event{"1": event}}
	server := NewServer(&mockLogger{}, mockApp, "localhost", "8080")

	req := httptest.NewRequest(http.MethodGet, "/api/events/day?date="+dateStr, nil)
	w := httptest.NewRecorder()

	server.mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp map[string][]eventDTO
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}
	if len(resp["events"]) != 1 {
		t.Errorf("expected 1 event, got %d", len(resp["events"]))
	}
}

func TestServer_StorageError(t *testing.T) {
	mockApp := &mockApplication{err: fmt.Errorf("some database error")}
	server := NewServer(&mockLogger{}, mockApp, "localhost", "8080")

	req := httptest.NewRequest(http.MethodGet, "/api/events/123", nil)
	w := httptest.NewRecorder()

	server.mux.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", w.Code)
	}
}
