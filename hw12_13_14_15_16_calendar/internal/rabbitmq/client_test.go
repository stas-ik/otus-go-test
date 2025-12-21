package rabbitmq

import (
	"encoding/json"
	"testing"
	"time"
)

func TestNotificationSerialization(t *testing.T) {
	now := time.Now().Round(time.Second)
	n := Notification{
		EventID:   "123",
		Title:     "Test Event",
		StartTime: now,
		UserID:    "user1",
	}

	data, err := json.Marshal(n)
	if err != nil {
		t.Fatalf("failed to marshal notification: %v", err)
	}

	var n2 Notification
	if err := json.Unmarshal(data, &n2); err != nil {
		t.Fatalf("failed to unmarshal notification: %v", err)
	}

	if n2.EventID != n.EventID {
		t.Errorf("expected EventID %s, got %s", n.EventID, n2.EventID)
	}
	if n2.Title != n.Title {
		t.Errorf("expected Title %s, got %s", n.Title, n2.Title)
	}
	if !n2.StartTime.Equal(n.StartTime) {
		t.Errorf("expected StartTime %v, got %v", n.StartTime, n2.StartTime)
	}
	if n2.UserID != n.UserID {
		t.Errorf("expected UserID %s, got %s", n.UserID, n2.UserID)
	}
}
