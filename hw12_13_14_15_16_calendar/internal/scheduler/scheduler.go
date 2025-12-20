package scheduler

import (
	"context"
	"fmt"
	"time"

	"github.com/stas-ik/otus-go-test/hw12_13_14_15_16_calendar/internal/logger"
	"github.com/stas-ik/otus-go-test/hw12_13_14_15_16_calendar/internal/rabbitmq"
	"github.com/stas-ik/otus-go-test/hw12_13_14_15_16_calendar/internal/storage"
)

type Storage interface {
	GetEventsToNotify(ctx context.Context) ([]storage.Event, error)
	MarkEventNotified(ctx context.Context, id string) error
	DeleteOldEvents(ctx context.Context, olderThan time.Time) error
}

type Publisher interface {
	Publish(n rabbitmq.Notification) error
}

type Scheduler struct {
	storage   Storage
	publisher Publisher
	logger    *logger.Logger
}

func New(s Storage, p Publisher, l *logger.Logger) *Scheduler {
	return &Scheduler{
		storage:   s,
		publisher: p,
		logger:    l,
	}
}

func (s *Scheduler) ProcessNotifications(ctx context.Context) {
	events, err := s.storage.GetEventsToNotify(ctx)
	if err != nil {
		s.logger.Error(fmt.Sprintf("failed to get events to notify: %v", err))
		return
	}

	for _, e := range events {
		notif := rabbitmq.Notification{
			EventID:   e.ID,
			Title:     e.Title,
			StartTime: e.StartTime,
			UserID:    e.UserID,
		}

		if err := s.publisher.Publish(notif); err != nil {
			s.logger.Error(fmt.Sprintf("failed to publish notification for event %s: %v", e.ID, err))
			continue
		}

		if err := s.storage.MarkEventNotified(ctx, e.ID); err != nil {
			s.logger.Error(fmt.Sprintf("failed to mark event %s as notified: %v", e.ID, err))
		} else {
			s.logger.Info(fmt.Sprintf("notification sent for event %s", e.ID))
		}
	}
}

func (s *Scheduler) ProcessCleanup(ctx context.Context) {
	olderThan := time.Now().AddDate(-1, 0, 0)
	if err := s.storage.DeleteOldEvents(ctx, olderThan); err != nil {
		s.logger.Error(fmt.Sprintf("failed to cleanup old events: %v", err))
	} else {
		s.logger.Info("old events cleaned up")
	}
}
