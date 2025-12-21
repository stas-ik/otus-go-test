package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"os"
	"testing"
	"time"
)

var calendarURL = "http://calendar:8080/api"

func TestMain(m *testing.M) {
	if url := os.Getenv("CALENDAR_URL"); url != "" {
		calendarURL = url
	}
	os.Exit(m.Run())
}

type eventDTO struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	StartTime   string `json:"startTime"`
	EndTime     string `json:"endTime"`
	Description string `json:"description"`
	UserID      string `json:"userId"`
}

func createEvent(ctx context.Context, t *testing.T, event eventDTO) {
	t.Helper()
	body, _ := json.Marshal(event)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, calendarURL+"/events", bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("failed to create event: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201 Created, got %d", resp.StatusCode)
	}
}

func getEvent(ctx context.Context, t *testing.T, id string) eventDTO {
	t.Helper()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, calendarURL+"/events/"+id, nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("failed to get event: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", resp.StatusCode)
	}
	var fetched eventDTO
	if err := json.NewDecoder(resp.Body).Decode(&fetched); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	return fetched
}

func listEvents(ctx context.Context, t *testing.T, date string) []eventDTO {
	t.Helper()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, calendarURL+"/events/day?date="+date, nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("failed to list events: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", resp.StatusCode)
	}
	var list []eventDTO
	if err := json.NewDecoder(resp.Body).Decode(&list); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	return list
}

func TestIntegration(t *testing.T) {
	ctx := context.Background()
	t.Run("Create and List Events", func(t *testing.T) {
		event := eventDTO{
			ID:          "test-event-1",
			Title:       "Integration Test Event",
			StartTime:   time.Now().Add(time.Hour).Format("2006-01-02 15:04:05"),
			EndTime:     time.Now().Add(time.Hour * 2).Format("2006-01-02 15:04:05"),
			Description: "Test Description",
			UserID:      "user1",
		}

		// 1. Create Event
		createEvent(ctx, t, event)

		// 2. Get Event by ID
		fetched := getEvent(ctx, t, event.ID)
		if fetched.ID != event.ID {
			t.Errorf("expected ID %s, got %s", event.ID, fetched.ID)
		}

		// 3. List Events for Day
		date := time.Now().Format("2006-01-02")
		list := listEvents(ctx, t, date)
		found := false
		for _, e := range list {
			if e.ID == event.ID {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("event not found in day list")
		}
	})

	t.Run("Business Errors - Time Busy", func(t *testing.T) {
		startTime := time.Now().Add(time.Hour * 10).Format("2006-01-02 15:04:05")
		endTime := time.Now().Add(time.Hour * 11).Format("2006-01-02 15:04:05")

		event1 := eventDTO{
			ID:        "busy-1",
			Title:     "Event 1",
			StartTime: startTime,
			EndTime:   endTime,
			UserID:    "user-busy",
		}
		event2 := eventDTO{
			ID:        "busy-2",
			Title:     "Event 2",
			StartTime: startTime,
			EndTime:   endTime,
			UserID:    "user-busy",
		}

		// Create first event
		createEvent(ctx, t, event1)

		// Try to create second event in the same time
		body2, _ := json.Marshal(event2)
		req2, _ := http.NewRequestWithContext(ctx, http.MethodPost, calendarURL+"/events", bytes.NewBuffer(body2))
		req2.Header.Set("Content-Type", "application/json")
		resp, err := http.DefaultClient.Do(req2)
		if err != nil {
			t.Fatalf("failed to post second event: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusConflict {
			t.Errorf("expected 409 Conflict for overlapping events, got %d", resp.StatusCode)
		}
	})
}
