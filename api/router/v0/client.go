package v0

import (
	"context"
	"time"

	"github.com/jgfranco17/echoris/api/events"
	pb "github.com/jgfranco17/echoris/service/protos"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// gRPC client connection (can reuse)
var grpcConn *grpc.ClientConn
var grpcClient pb.LogAggregatorClient

func initGRPC() error {
	var err error
	creds := grpc.WithTransportCredentials(insecure.NewCredentials())
	grpcConn, err = grpc.NewClient("worker-service:50051", creds)
	if err != nil {
		return err
	}
	grpcClient = pb.NewLogAggregatorClient(grpcConn)
	return nil
}

func forwardLogs(batch []events.Entry) error {
	if grpcClient == nil {
		if err := initGRPC(); err != nil {
			return err
		}
	}

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

	_, err := grpcClient.SendLogs(context.Background(), &pb.LogBatch{
		Events: pbEvents,
	})
	return err
}

func fetchLogs(service, level string) ([]events.Entry, error) {
	if grpcClient == nil {
		if err := initGRPC(); err != nil {
			return nil, err
		}
	}

	resp, err := grpcClient.QueryLogs(context.Background(), &pb.QueryRequest{
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
