//go:build grpcapi

package grpcserver

import (
	"context"
	"log"
	"net"
	"testing"
	"time"

	"github.com/stas-ik/otus-go-test/hw12_13_14_15_16_calendar/api/gen"
	"github.com/stas-ik/otus-go-test/hw12_13_14_15_16_calendar/internal/storage"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const bufSize = 1024 * 1024

var lis *bufconn.Listener

type mockApplication struct {
	events map[string]storage.Event
	err    error
}

func (m *mockApplication) CreateEvent(ctx context.Context, event storage.Event) error {
	if m.err != nil {
		return m.err
	}
	m.events[event.ID] = event
	return nil
}

func (m *mockApplication) UpdateEvent(ctx context.Context, id string, event storage.Event) error {
	if m.err != nil {
		return m.err
	}
	m.events[id] = event
	return nil
}

func (m *mockApplication) DeleteEvent(ctx context.Context, id string) error {
	if m.err != nil {
		return m.err
	}
	delete(m.events, id)
	return nil
}

func (m *mockApplication) GetEventByID(ctx context.Context, id string) (*storage.Event, error) {
	if m.err != nil {
		return nil, m.err
	}
	ev, ok := m.events[id]
	if !ok {
		return nil, storage.ErrEventNotFound
	}
	return &ev, nil
}

func (m *mockApplication) ListEventsForDay(ctx context.Context, date time.Time) ([]storage.Event, error) {
	if m.err != nil {
		return nil, m.err
	}
	return nil, nil
}

func (m *mockApplication) ListEventsForWeek(ctx context.Context, startDate time.Time) ([]storage.Event, error) {
	return nil, nil
}

func (m *mockApplication) ListEventsForMonth(ctx context.Context, startDate time.Time) ([]storage.Event, error) {
	return nil, nil
}

type mockLogger struct{}

func (m *mockLogger) Info(msg string)  {}
func (m *mockLogger) Error(msg string) {}

func dialer(ctx context.Context, s *grpc.Server) func(context.Context, string) (net.Conn, error) {
	lis = bufconn.Listen(bufSize)
	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatalf("Server exited with error: %v", err)
		}
	}()
	return func(context.Context, string) (net.Conn, error) {
		return lis.Dial()
	}
}

func TestGRPCServer_CreateEvent(t *testing.T) {
	ctx := context.Background()
	mockApp := &mockApplication{events: make(map[string]storage.Event)}

	s := grpc.NewServer()
	srv := &Server{
		app:    mockApp,
		logger: &mockLogger{},
		srv:    s,
	}
	gen.RegisterCalendarServiceServer(s, srv)

	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(dialer(ctx, s)), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}
	defer conn.Close()
	client := gen.NewCalendarServiceClient(conn)

	event := &gen.Event{
		Id:        "1",
		Title:     "GRPC Event",
		StartTime: timestamppb.Now(),
		EndTime:   timestamppb.Now(),
		UserId:    "user1",
	}

	_, err = client.CreateEvent(ctx, &gen.CreateEventRequest{Event: event})
	if err != nil {
		t.Fatalf("CreateEvent failed: %v", err)
	}

	if len(mockApp.events) != 1 {
		t.Errorf("expected 1 event, got %d", len(mockApp.events))
	}

	s.Stop()
}

func TestGRPCServer_GetEventByID(t *testing.T) {
	ctx := context.Background()
	id := "123"
	mockApp := &mockApplication{
		events: map[string]storage.Event{
			id: {ID: id, Title: "Test Event", StartTime: time.Now(), EndTime: time.Now()},
		},
	}

	s := grpc.NewServer()
	srv := &Server{
		app:    mockApp,
		logger: &mockLogger{},
		srv:    s,
	}
	gen.RegisterCalendarServiceServer(s, srv)

	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(dialer(ctx, s)), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}
	defer conn.Close()
	client := gen.NewCalendarServiceClient(conn)

	resp, err := client.GetEventByID(ctx, &gen.GetEventByIDRequest{Id: id})
	if err != nil {
		t.Fatalf("GetEventByID failed: %v", err)
	}

	if resp.GetEvent().GetTitle() != "Test Event" {
		t.Errorf("expected title Test Event, got %s", resp.GetEvent().GetTitle())
	}

	s.Stop()
}

func TestGRPCServer_DeleteEvent(t *testing.T) {
	ctx := context.Background()
	id := "123"
	mockApp := &mockApplication{
		events: map[string]storage.Event{
			id: {ID: id},
		},
	}

	s := grpc.NewServer()
	srv := &Server{
		app:    mockApp,
		logger: &mockLogger{},
		srv:    s,
	}
	gen.RegisterCalendarServiceServer(s, srv)

	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(dialer(ctx, s)), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}
	defer conn.Close()
	client := gen.NewCalendarServiceClient(conn)

	_, err = client.DeleteEvent(ctx, &gen.DeleteEventRequest{Id: id})
	if err != nil {
		t.Fatalf("DeleteEvent failed: %v", err)
	}

	if len(mockApp.events) != 0 {
		t.Errorf("expected 0 events, got %d", len(mockApp.events))
	}

	s.Stop()
}
