package v0

import (
	"context"
	"time"

	"github.com/jgfranco17/echoris/api/events"
	pb "github.com/jgfranco17/echoris/service/protos"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// LogClient defines the interface for log operations
type LogClient interface {
	ForwardLogs(ctx context.Context, batch []events.Entry) error
	FetchLogs(ctx context.Context, service, level string) ([]events.Entry, error)
	Close() error
}

// GRPCLogClient implements LogClient using gRPC
type GRPCLogClient struct {
	conn   *grpc.ClientConn
	client pb.LogAggregatorClient
}

// NewGRPCLogClient creates a new gRPC log client
func NewGRPCLogClient(address string) (*GRPCLogClient, error) {
	creds := grpc.WithTransportCredentials(insecure.NewCredentials())
	conn, err := grpc.NewClient(address, creds)
	if err != nil {
		return nil, err
	}

	client := pb.NewLogAggregatorClient(conn)

	return &GRPCLogClient{
		conn:   conn,
		client: client,
	}, nil
}

// ForwardLogs sends a batch of log entries to the log aggregator
func (c *GRPCLogClient) ForwardLogs(ctx context.Context, batch []events.Entry) error {
	var pbEvents []*pb.LogEvent
	for _, e := range batch {
		pbEvents = append(pbEvents, &pb.LogEvent{
			Timestamp: e.Timestamp.Format(time.RFC3339Nano),
			Service:   e.Service,
			Level:     e.Level,
			Message:   e.Message,
			Fields:    e.Fields,
		})
	}

	_, err := c.client.SendLogs(ctx, &pb.LogBatch{
		Events: pbEvents,
	})
	return err
}

// FetchLogs retrieves log entries from the log aggregator
func (c *GRPCLogClient) FetchLogs(ctx context.Context, service, level string) ([]events.Entry, error) {
	resp, err := c.client.QueryLogs(ctx, &pb.QueryRequest{
		Service: service,
		Level:   level,
	})
	if err != nil {
		return nil, err
	}

	var logs []events.Entry
	for _, e := range resp.Events {
		t, _ := time.Parse(time.RFC3339Nano, e.Timestamp)
		logs = append(logs, events.Entry{
			Timestamp: t,
			Service:   e.Service,
			Level:     e.Level,
			Message:   e.Message,
			Fields:    e.Fields,
		})
	}

	return logs, nil
}

// Close closes the gRPC connection
func (c *GRPCLogClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
