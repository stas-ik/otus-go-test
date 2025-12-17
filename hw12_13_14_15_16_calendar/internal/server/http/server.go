package internalhttp

import (
	"context"
	"fmt"
	"net/http"
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

// Application — заглушка под бизнес-логику. На данном этапе сервер не использует её напрямую,
// но зависимость оставлена для дальнейшего развития.
type Application interface{}

// NewServer конструирует HTTP-сервер с hello endpoint и логирующей middleware.
func NewServer(logger Logger, app Application, host, port string) *Server {
	mux := http.NewServeMux()

	// Hello endpoint
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("hello world"))
	})
	mux.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("hello world"))
	})

	// Оборачиваем middleware для логирования запросов
	handler := loggingMiddleware(logger)(mux)

	srv := &http.Server{
		Addr:    fmt.Sprintf("%s:%s", host, port),
		Handler: handler,
	}

	return &Server{
		logger: logger,
		app:    app,
		srv:    srv,
		mux:    mux,
	}
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
