//go:build grpcapi

package grpcserver

import (
	"context"
	"fmt"
	"net"
	"time"

	gen "github.com/stas-ik/otus-go-test/hw12_13_14_15_16_calendar/api/gen"
	"github.com/stas-ik/otus-go-test/hw12_13_14_15_16_calendar/internal/storage"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Logger минимальный интерфейс логгера.
type Logger interface {
	Info(msg string)
	Error(msg string)
}

// Application - интерфейс доменной логики (совпадает с HTTP слоем).
type Application interface {
	CreateEvent(ctx context.Context, event storage.Event) error
	UpdateEvent(ctx context.Context, id string, event storage.Event) error
	DeleteEvent(ctx context.Context, id string) error
	GetEventByID(ctx context.Context, id string) (*storage.Event, error)
	ListEventsForDay(ctx context.Context, date time.Time) ([]storage.Event, error)
	ListEventsForWeek(ctx context.Context, startDate time.Time) ([]storage.Event, error)
	ListEventsForMonth(ctx context.Context, startDate time.Time) ([]storage.Event, error)
}

type Server struct {
	gen.UnimplementedCalendarServiceServer
	logger Logger
	app    Application
	srv    *grpc.Server
	lis    net.Listener
}

func New(logger Logger, app Application, host, port string) (*Server, error) {
	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%s", host, port))
	if err != nil {
		return nil, err
	}
	s := &Server{
		logger: logger,
		app:    app,
		srv:    grpc.NewServer(),
		lis:    lis,
	}
	gen.RegisterCalendarServiceServer(s.srv, s)
	return s, nil
}

func (s *Server) Start() error {
	return s.srv.Serve(s.lis)
}

func (s *Server) Stop() {
	s.srv.GracefulStop()
}

// ===== RPC handlers =====

func (s *Server) CreateEvent(ctx context.Context, req *gen.CreateEventRequest) (*gen.CreateEventResponse, error) {
	ev := fromPB(req.GetEvent())
	return &gen.CreateEventResponse{}, s.app.CreateEvent(ctx, ev)
}

func (s *Server) UpdateEvent(ctx context.Context, req *gen.UpdateEventRequest) (*gen.UpdateEventResponse, error) {
	ev := fromPB(req.GetEvent())
	return &gen.UpdateEventResponse{}, s.app.UpdateEvent(ctx, req.GetId(), ev)
}

func (s *Server) DeleteEvent(ctx context.Context, req *gen.DeleteEventRequest) (*gen.DeleteEventResponse, error) {
	return &gen.DeleteEventResponse{}, s.app.DeleteEvent(ctx, req.GetId())
}

func (s *Server) GetEventByID(ctx context.Context, req *gen.GetEventByIDRequest) (*gen.GetEventByIDResponse, error) {
	ev, err := s.app.GetEventByID(ctx, req.GetId())
	if err != nil {
		return nil, err
	}
	return &gen.GetEventByIDResponse{Event: toPB(*ev)}, nil
}

func (s *Server) ListEventsForDay(ctx context.Context, req *gen.ListForDayRequest) (*gen.ListEventsResponse, error) {
	date := req.GetDate().AsTime()
	list, err := s.app.ListEventsForDay(ctx, date)
	if err != nil {
		return nil, err
	}
	return &gen.ListEventsResponse{Events: toPBList(list)}, nil
}

func (s *Server) ListEventsForWeek(ctx context.Context, req *gen.ListForWeekRequest) (*gen.ListEventsResponse, error) {
	start := req.GetStart().AsTime()
	list, err := s.app.ListEventsForWeek(ctx, start)
	if err != nil {
		return nil, err
	}
	return &gen.ListEventsResponse{Events: toPBList(list)}, nil
}

func (s *Server) ListEventsForMonth(ctx context.Context, req *gen.ListForMonthRequest) (*gen.ListEventsResponse, error) {
	start := req.GetStart().AsTime()
	list, err := s.app.ListEventsForMonth(ctx, start)
	if err != nil {
		return nil, err
	}
	return &gen.ListEventsResponse{Events: toPBList(list)}, nil
}

// ===== mapping =====

func toPB(e storage.Event) *gen.Event {
	var notify *timestamppb.Timestamp
	if e.NotifyAt != nil {
		notify = timestamppb.New(*e.NotifyAt)
	}
	return &gen.Event{
		Id:          e.ID,
		Title:       e.Title,
		StartTime:   timestamppb.New(e.StartTime),
		EndTime:     timestamppb.New(e.EndTime),
		Description: e.Description,
		UserId:      e.UserID,
		NotifyAt:    notify,
	}
}

func fromPB(e *gen.Event) storage.Event {
	var notify *time.Time
	if e.GetNotifyAt() != nil {
		t := e.GetNotifyAt().AsTime()
		notify = &t
	}
	return storage.Event{
		ID:          e.GetId(),
		Title:       e.GetTitle(),
		StartTime:   e.GetStartTime().AsTime(),
		EndTime:     e.GetEndTime().AsTime(),
		Description: e.GetDescription(),
		UserID:      e.GetUserId(),
		NotifyAt:    notify,
	}
}

func toPBList(list []storage.Event) []*gen.Event {
	res := make([]*gen.Event, 0, len(list))
	for _, e := range list {
		res = append(res, toPB(e))
	}
	return res
}
