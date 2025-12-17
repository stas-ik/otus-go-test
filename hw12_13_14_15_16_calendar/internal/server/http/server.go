package internalhttp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/stas-ik/otus-go-test/hw12_13_14_15_16_calendar/internal/storage" //nolint:depguard
)

type Server struct {
	logger Logger
	app    Application
	srv    *http.Server
	mux    *http.ServeMux
}

// Logger описывает минимальные методы логгера, используемые сервером и middleware.
type Logger interface {
	Info(msg string)
	Error(msg string)
}

// Application — интерфейс доменной логики, используемый HTTP-слоем.
type Application interface {
	CreateEvent(ctx context.Context, event storage.Event) error
	UpdateEvent(ctx context.Context, id string, event storage.Event) error
	DeleteEvent(ctx context.Context, id string) error
	GetEventByID(ctx context.Context, id string) (*storage.Event, error)
	ListEventsForDay(ctx context.Context, date time.Time) ([]storage.Event, error)
	ListEventsForWeek(ctx context.Context, startDate time.Time) ([]storage.Event, error)
	ListEventsForMonth(ctx context.Context, startDate time.Time) ([]storage.Event, error)
}

// NewServer конструирует HTTP-сервер с hello endpoint и логирующей middleware.
func NewServer(logger Logger, app Application, host, port string) *Server {
	mux := http.NewServeMux()

	s := &Server{
		logger: logger,
		app:    app,
		mux:    mux,
	}

	s.registerRoutes(mux)

	// Оборачиваем middleware для логирования запросов
	handler := loggingMiddleware(logger)(mux)

	srv := &http.Server{
		Addr:              fmt.Sprintf("%s:%s", host, port),
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
	}

	s.srv = srv
	return s
}

// Start запускает HTTP-сервер и блокируется до завершения контекста или возникновения ошибки запуска.
func (s *Server) Start(ctx context.Context) error {
	errCh := make(chan error, 1)
	go func() {
		if err := s.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
		close(errCh)
	}()

	select {
	case <-ctx.Done():
		return nil
	case err := <-errCh:
		return err
	}
}

// Stop выполняет graceful shutdown с использованием переданного контекста.
func (s *Server) Stop(ctx context.Context) error {
	return s.srv.Shutdown(ctx)
}

// ===== HTTP API =====

type eventDTO struct {
	ID          string     `json:"id,omitempty"`
	Title       string     `json:"title"`
	StartTime   time.Time  `json:"startTime"`
	EndTime     time.Time  `json:"endTime"`
	Description string     `json:"description,omitempty"`
	UserID      string     `json:"userId"`
	NotifyAt    *time.Time `json:"notifyAt,omitempty"`
}

func toDTO(e storage.Event) eventDTO {
	return eventDTO{
		ID:          e.ID,
		Title:       e.Title,
		StartTime:   e.StartTime,
		EndTime:     e.EndTime,
		Description: e.Description,
		UserID:      e.UserID,
		NotifyAt:    e.NotifyAt,
	}
}

func fromDTO(d eventDTO) storage.Event {
	return storage.Event{
		ID:          d.ID,
		Title:       d.Title,
		StartTime:   d.StartTime,
		EndTime:     d.EndTime,
		Description: d.Description,
		UserID:      d.UserID,
		NotifyAt:    d.NotifyAt,
	}
}

func (s *Server) registerRoutes(mux *http.ServeMux) {
	// health/hello
	mux.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("hello world"))
	})
	mux.HandleFunc("/hello", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("hello world"))
	})

	// CRUD для событий
	mux.HandleFunc("/api/events", s.handleEventsRoot)
	mux.HandleFunc("/api/events/", s.handleEventByID)

	// Листинги по интервалам
	mux.HandleFunc("/api/events/day", s.handleListDay)
	mux.HandleFunc("/api/events/week", s.handleListWeek)
	mux.HandleFunc("/api/events/month", s.handleListMonth)
}

func (s *Server) handleEventsRoot(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		var d eventDTO
		if err := json.NewDecoder(r.Body).Decode(&d); err != nil {
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}
		ev := fromDTO(d)
		if err := s.app.CreateEvent(r.Context(), ev); err != nil {
			s.writeStorageError(w, err)
			return
		}
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "created"})
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleEventByID(w http.ResponseWriter, r *http.Request) {
	// URL: /api/events/{id}
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/events/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		http.NotFound(w, r)
		return
	}
	id := parts[0]

	switch r.Method {
	case http.MethodGet:
		ev, err := s.app.GetEventByID(r.Context(), id)
		if err != nil {
			s.writeStorageError(w, err)
			return
		}
		_ = json.NewEncoder(w).Encode(toDTO(*ev))
	case http.MethodPut:
		var d eventDTO
		if err := json.NewDecoder(r.Body).Decode(&d); err != nil {
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}
		if err := s.app.UpdateEvent(r.Context(), id, fromDTO(d)); err != nil {
			s.writeStorageError(w, err)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "updated"})
	case http.MethodDelete:
		if err := s.app.DeleteEvent(r.Context(), id); err != nil {
			s.writeStorageError(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleListDay(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	dateStr := r.URL.Query().Get("date")
	if dateStr == "" {
		http.Error(w, "missing date", http.StatusBadRequest)
		return
	}
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		http.Error(w, "invalid date format, want YYYY-MM-DD", http.StatusBadRequest)
		return
	}
	list, err := s.app.ListEventsForDay(r.Context(), date)
	if err != nil {
		s.writeStorageError(w, err)
		return
	}
	_ = json.NewEncoder(w).Encode(map[string][]eventDTO{"events": toDTOList(list)})
}

func (s *Server) handleListWeek(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	startStr := r.URL.Query().Get("start")
	if startStr == "" {
		http.Error(w, "missing start", http.StatusBadRequest)
		return
	}
	start, err := time.Parse("2006-01-02", startStr)
	if err != nil {
		http.Error(w, "invalid start format, want YYYY-MM-DD", http.StatusBadRequest)
		return
	}
	list, err := s.app.ListEventsForWeek(r.Context(), start)
	if err != nil {
		s.writeStorageError(w, err)
		return
	}
	_ = json.NewEncoder(w).Encode(map[string][]eventDTO{"events": toDTOList(list)})
}

func (s *Server) handleListMonth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	startStr := r.URL.Query().Get("start")
	if startStr == "" {
		http.Error(w, "missing start", http.StatusBadRequest)
		return
	}
	start, err := time.Parse("2006-01-02", startStr)
	if err != nil {
		http.Error(w, "invalid start format, want YYYY-MM-DD", http.StatusBadRequest)
		return
	}
	list, err := s.app.ListEventsForMonth(r.Context(), start)
	if err != nil {
		s.writeStorageError(w, err)
		return
	}
	_ = json.NewEncoder(w).Encode(map[string][]eventDTO{"events": toDTOList(list)})
}

func toDTOList(list []storage.Event) []eventDTO {
	res := make([]eventDTO, 0, len(list))
	for _, e := range list {
		res = append(res, toDTO(e))
	}
	return res
}

func (s *Server) writeStorageError(w http.ResponseWriter, err error) {
	var status int
	switch {
	case errors.Is(err, storage.ErrInvalidEvent):
		status = http.StatusBadRequest
	case errors.Is(err, storage.ErrEventNotFound):
		status = http.StatusNotFound
	case errors.Is(err, storage.ErrDateBusy):
		status = http.StatusConflict
	default:
		status = http.StatusInternalServerError
	}
	http.Error(w, err.Error(), status)
}
