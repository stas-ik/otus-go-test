package sqlstorage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // register postgres driver for database/sql
	"github.com/stas-ik/otus-go-test/hw12_13_14_15_16_calendar/internal/storage"
)

type Storage struct {
	dsn string
	db  *sqlx.DB
}

func New(dsn string) *Storage {
	return &Storage{
		dsn: dsn,
	}
}

func (s *Storage) Connect(ctx context.Context) error {
	db, err := sqlx.ConnectContext(ctx, "postgres", s.dsn)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	s.db = db
	return nil
}

func (s *Storage) Close(_ context.Context) error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

func (s *Storage) CreateEvent(ctx context.Context, event storage.Event) error {
	if event.ID == "" || event.Title == "" {
		return storage.ErrInvalidEvent
	}

	busy, err := s.isTimeBusy(ctx, "", event.UserID, event.StartTime, event.EndTime)
	if err != nil {
		return fmt.Errorf("failed to check if time is busy: %w", err)
	}
	if busy {
		return storage.ErrDateBusy
	}

	query := `
		INSERT INTO events (id, title, start_time, end_time, description, user_id, notify_at, notified)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err = s.db.ExecContext(ctx, query,
		event.ID,
		event.Title,
		event.StartTime,
		event.EndTime,
		event.Description,
		event.UserID,
		event.NotifyAt,
		event.Notified,
	)
	if err != nil {
		return fmt.Errorf("failed to create event: %w", err)
	}

	return nil
}

func (s *Storage) UpdateEvent(ctx context.Context, id string, event storage.Event) error {
	if event.Title == "" {
		return storage.ErrInvalidEvent
	}

	var exists bool
	err := s.db.GetContext(ctx, &exists, "SELECT EXISTS(SELECT 1 FROM events WHERE id = $1)", id)
	if err != nil {
		return fmt.Errorf("failed to check event existence: %w", err)
	}
	if !exists {
		return storage.ErrEventNotFound
	}

	busy, err := s.isTimeBusy(ctx, id, event.UserID, event.StartTime, event.EndTime)
	if err != nil {
		return fmt.Errorf("failed to check if time is busy: %w", err)
	}
	if busy {
		return storage.ErrDateBusy
	}

	query := `
		UPDATE events
		SET title = $2, start_time = $3, end_time = $4, description = $5, user_id = $6, notify_at = $7, notified = $8
		WHERE id = $1
	`

	_, err = s.db.ExecContext(ctx, query,
		id,
		event.Title,
		event.StartTime,
		event.EndTime,
		event.Description,
		event.UserID,
		event.NotifyAt,
		event.Notified,
	)
	if err != nil {
		return fmt.Errorf("failed to update event: %w", err)
	}

	return nil
}

func (s *Storage) DeleteEvent(ctx context.Context, id string) error {
	result, err := s.db.ExecContext(ctx, "DELETE FROM events WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to delete event: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rows == 0 {
		return storage.ErrEventNotFound
	}

	return nil
}

func (s *Storage) GetEventByID(ctx context.Context, id string) (*storage.Event, error) {
	var event storage.Event

	query := `
		SELECT id, title, start_time, end_time, description, user_id, notify_at, notified
		FROM events
		WHERE id = $1
	`

	err := s.db.GetContext(ctx, &event, query, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, storage.ErrEventNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get event: %w", err)
	}

	return &event, nil
}

func (s *Storage) ListEventsForDay(ctx context.Context, date time.Time) ([]storage.Event, error) {
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	return s.listEventsBetween(ctx, startOfDay, endOfDay)
}

func (s *Storage) ListEventsForWeek(ctx context.Context, startDate time.Time) ([]storage.Event, error) {
	startOfWeek := time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, startDate.Location())
	endOfWeek := startOfWeek.Add(7 * 24 * time.Hour)

	return s.listEventsBetween(ctx, startOfWeek, endOfWeek)
}

func (s *Storage) ListEventsForMonth(ctx context.Context, startDate time.Time) ([]storage.Event, error) {
	startOfMonth := time.Date(startDate.Year(), startDate.Month(), 1, 0, 0, 0, 0, startDate.Location())
	endOfMonth := startOfMonth.AddDate(0, 1, 0)

	return s.listEventsBetween(ctx, startOfMonth, endOfMonth)
}

func (s *Storage) listEventsBetween(ctx context.Context, start, end time.Time) ([]storage.Event, error) {
	var events []storage.Event

	query := `
		SELECT id, title, start_time, end_time, description, user_id, notify_at, notified
		FROM events
		WHERE start_time >= $1 AND start_time < $2
		ORDER BY start_time
	`

	err := s.db.SelectContext(ctx, &events, query, start, end)
	if err != nil {
		return nil, fmt.Errorf("failed to list events: %w", err)
	}

	if events == nil {
		events = []storage.Event{}
	}

	return events, nil
}

func (s *Storage) isTimeBusy(ctx context.Context, excludeID, userID string, start, end time.Time) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM events
			WHERE user_id = $1
			AND id != $2
			AND start_time < $4
			AND end_time > $3
		)
	`

	var busy bool
	err := s.db.GetContext(ctx, &busy, query, userID, excludeID, start, end)
	if err != nil {
		return false, err
	}

	return busy, nil
}

func (s *Storage) GetEventsToNotify(ctx context.Context) ([]storage.Event, error) {
	var events []storage.Event

	query := `
		SELECT id, title, start_time, end_time, description, user_id, notify_at, notified
		FROM events
		WHERE notified = FALSE AND notify_at <= NOW()
	`

	err := s.db.SelectContext(ctx, &events, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get events to notify: %w", err)
	}

	if events == nil {
		events = []storage.Event{}
	}

	return events, nil
}

func (s *Storage) MarkEventNotified(ctx context.Context, id string) error {
	query := `UPDATE events SET notified = TRUE WHERE id = $1`
	_, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to mark event as notified: %w", err)
	}
	return nil
}

func (s *Storage) DeleteOldEvents(ctx context.Context, olderThan time.Time) error {
	query := `DELETE FROM events WHERE start_time < $1`
	_, err := s.db.ExecContext(ctx, query, olderThan)
	if err != nil {
		return fmt.Errorf("failed to delete old events: %w", err)
	}
	return nil
}
