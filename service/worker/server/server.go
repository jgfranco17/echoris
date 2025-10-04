package server

import (
	"context"
	"time"

	pb "github.com/jgfranco17/echoris/service/protos"
	"github.com/jgfranco17/echoris/service/worker/storage"
	"github.com/sirupsen/logrus"
)

// LogAggregatorServer implements the LogAggregator gRPC service
type LogAggregatorServer struct {
	pb.UnimplementedLogAggregatorServer
	storage storage.Storage
	logger  *logrus.Logger
}

// NewLogAggregatorServer creates a new LogAggregatorServer instance
func NewLogAggregatorServer(store storage.Storage, logger *logrus.Logger) *LogAggregatorServer {
	return &LogAggregatorServer{
		storage: store,
		logger:  logger,
	}
}

// SendLogs receives and stores a batch of log events
func (s *LogAggregatorServer) SendLogs(ctx context.Context, batch *pb.LogBatch) (*pb.SendLogsResponse, error) {
	if len(batch.Events) == 0 {
		return &pb.SendLogsResponse{Ok: true}, nil
	}

	s.logger.WithField("count", len(batch.Events)).Info("Receiving log batch")

	entries := make([]storage.LogEntry, 0, len(batch.Events))
	for _, e := range batch.Events {
		t, err := time.Parse(time.RFC3339Nano, e.Timestamp)
		if err != nil {
			s.logger.WithError(err).Warn("Failed to parse timestamp, using current time")
			t = time.Now().UTC()
		}

		entries = append(entries, storage.LogEntry{
			Timestamp: t,
			Service:   e.Service,
			Level:     e.Level,
			Message:   e.Message,
			Fields:    e.Fields,
		})
	}

	if err := s.storage.InsertLogs(ctx, entries); err != nil {
		s.logger.WithError(err).Error("Failed to insert logs")
		return &pb.SendLogsResponse{Ok: false}, err
	}

	s.logger.WithField("count", len(entries)).Info("Successfully stored logs")
	return &pb.SendLogsResponse{Ok: true}, nil
}

// QueryLogs retrieves logs based on the provided filters
func (s *LogAggregatorServer) QueryLogs(ctx context.Context, req *pb.QueryRequest) (*pb.LogBatch, error) {
	s.logger.WithFields(logrus.Fields{
		"service": req.Service,
		"level":   req.Level,
	}).Info("Querying logs")

	filter := storage.QueryFilter{
		Service: req.Service,
		Level:   req.Level,
		Limit:   1000, // Default limit to prevent large result sets
	}

	entries, err := s.storage.QueryLogs(ctx, filter)
	if err != nil {
		s.logger.WithError(err).Error("Failed to query logs")
		return nil, err
	}

	resp := &pb.LogBatch{
		Events: make([]*pb.LogEvent, 0, len(entries)),
	}

	for _, entry := range entries {
		resp.Events = append(resp.Events, &pb.LogEvent{
			Timestamp: entry.Timestamp.Format(time.RFC3339Nano),
			Service:   entry.Service,
			Level:     entry.Level,
			Message:   entry.Message,
			Fields:    entry.Fields,
		})
	}

	s.logger.WithField("count", len(resp.Events)).Info("Successfully retrieved logs")
	return resp, nil
}

// Close closes the server and its dependencies
func (s *LogAggregatorServer) Close() error {
	if s.storage != nil {
		return s.storage.Close()
	}
	return nil
}
