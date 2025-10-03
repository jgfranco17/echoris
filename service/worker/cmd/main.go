package main

import (
	"context"
	"sync"
	"time"

	pb "github.com/jgfranco17/echoris/service/protos"
)

type LogEvent struct {
	Timestamp time.Time
	Service   string
	Level     string
	Message   string
	Fields    map[string]string
}

type server struct {
	pb.UnimplementedLogAggregatorServer
	mu   sync.RWMutex
	logs []LogEvent
}

func (s *server) SendLogs(ctx context.Context, batch *pb.LogBatch) (*pb.SendLogsResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, e := range batch.Events {
		t, err := time.Parse(time.RFC3339Nano, e.Timestamp)
		if err != nil {
			t = time.Now().UTC()
		}
		s.logs = append(s.logs, LogEvent{
			Timestamp: t,
			Service:   e.Service,
			Level:     e.Level,
			Message:   e.Message,
			Fields:    e.Fields,
		})
	}
	return &pb.SendLogsResponse{Ok: true}, nil
}

func (s *server) QueryLogs(ctx context.Context, req *pb.QueryRequest) (*pb.LogBatch, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var resp pb.LogBatch
	for _, e := range s.logs {
		if req.Service != "" && e.Service != req.Service {
			continue
		}
		if req.Level != "" && e.Level != req.Level {
			continue
		}
		resp.Events = append(resp.Events, &pb.LogEvent{
			Timestamp: e.Timestamp.Format(time.RFC3339Nano),
			Service:   e.Service,
			Level:     e.Level,
			Message:   e.Message,
			Fields:    e.Fields,
		})
	}
	return &resp, nil
}
