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
	ID          string  `json:"id"`
	Title       string  `json:"title"`
	StartTime   string  `json:"startTime"`
	EndTime     string  `json:"endTime"`
	Description string  `json:"description"`
	UserID      string  `json:"userId"`
	NotifyAt    *string `json:"notifyAt,omitempty"`
	Notified    bool    `json:"notified,omitempty"`
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

func listEvents(ctx context.Context, t *testing.T, interval, date string) []eventDTO {
	t.Helper()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, calendarURL+"/events/"+interval+"?date="+date, nil)
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

func assertEventInList(t *testing.T, list []eventDTO, eventID string, listName string) {
	t.Helper()
	found := false
	for _, e := range list {
		if e.ID == eventID {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("%s event not found in %s list", eventID, listName)
	}
}

func testCreateAndListEvents(ctx context.Context, t *testing.T) {
	t.Helper()
	now := time.Now()
	eventDay := eventDTO{
		ID:        "test-day-1",
		Title:     "Day Event",
		StartTime: now.Add(time.Hour).Format("2006-01-02 15:04:05"),
		EndTime:   now.Add(time.Hour * 2).Format("2006-01-02 15:04:05"),
		UserID:    "user1",
	}
	eventWeek := eventDTO{
		ID:        "test-week-1",
		Title:     "Week Event",
		StartTime: now.AddDate(0, 0, 3).Format("2006-01-02 15:04:05"),
		EndTime:   now.AddDate(0, 0, 3).Add(time.Hour).Format("2006-01-02 15:04:05"),
		UserID:    "user1",
	}
	eventMonth := eventDTO{
		ID:        "test-month-1",
		Title:     "Month Event",
		StartTime: now.AddDate(0, 0, 20).Format("2006-01-02 15:04:05"),
		EndTime:   now.AddDate(0, 0, 20).Add(time.Hour).Format("2006-01-02 15:04:05"),
		UserID:    "user1",
	}

	createEvent(ctx, t, eventDay)
	createEvent(ctx, t, eventWeek)
	createEvent(ctx, t, eventMonth)

	// 1. List for Day
	list := listEvents(ctx, t, "day", now.Format("2006-01-02"))
	assertEventInList(t, list, eventDay.ID, "day")

	// 2. List for Week
	list = listEvents(ctx, t, "week", now.Format("2006-01-02"))
	assertEventInList(t, list, eventDay.ID, "week")
	assertEventInList(t, list, eventWeek.ID, "week")

	// 3. List for Month
	list = listEvents(ctx, t, "month", now.Format("2006-01-02"))
	assertEventInList(t, list, eventDay.ID, "month")
	assertEventInList(t, list, eventWeek.ID, "month")
	assertEventInList(t, list, eventMonth.ID, "month")
}

func testNotificationDelivery(ctx context.Context, t *testing.T) {
	t.Helper()
	// Создаем событие, которое должно быть отправлено прямо сейчас
	notifyTime := time.Now().Add(-time.Second).Format("2006-01-02 15:04:05")
	event := eventDTO{
		ID:        "notif-test-1",
		Title:     "Notify Me",
		StartTime: time.Now().Add(time.Hour).Format("2006-01-02 15:04:05"),
		EndTime:   time.Now().Add(time.Hour * 2).Format("2006-01-02 15:04:05"),
		UserID:    "user-notif",
		NotifyAt:  &notifyTime,
	}

	createEvent(ctx, t, event)

	// Ждем, пока планировщик и рассыльщик отработают
	// (интервал сканирования в конфиге 1м, но в тестах может быть меньше или мы просто подождем)
	// В deployments/config.docker.yaml scanInterval может быть маленьким.
	// Проверим статус notified через GetEventByID.

	maxWait := 15 * time.Second
	start := time.Now()
	success := false
	for time.Since(start) < maxWait {
		fetched := getEvent(ctx, t, event.ID)
		if fetched.Notified {
			success = true
			break
		}
		time.Sleep(1 * time.Second)
	}

	if !success {
		t.Errorf("event was not marked as notified within %v", maxWait)
	}
}

func testBusinessErrors(ctx context.Context, t *testing.T) {
	t.Helper()
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
}

func TestIntegration(t *testing.T) {
	ctx := context.Background()

	t.Run("Create and List Events", func(t *testing.T) {
		testCreateAndListEvents(ctx, t)
	})

	t.Run("Notification Delivery", func(t *testing.T) {
		testNotificationDelivery(ctx, t)
	})

	t.Run("Business Errors - Time Busy", func(t *testing.T) {
		testBusinessErrors(ctx, t)
	})
}
