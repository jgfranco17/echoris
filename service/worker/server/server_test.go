package server_test

import (
	"context"
	"io"
	"testing"
	"time"

	pb "github.com/jgfranco17/echoris/service/protos"
	"github.com/jgfranco17/echoris/service/worker/server"
	"github.com/jgfranco17/echoris/service/worker/storage"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockStorage is a mock implementation of storage.Storage
type MockStorage struct {
	mock.Mock
}

func (m *MockStorage) InsertLog(ctx context.Context, entry storage.LogEntry) error {
	args := m.Called(ctx, entry)
	return args.Error(0)
}

func (m *MockStorage) InsertLogs(ctx context.Context, entries []storage.LogEntry) error {
	args := m.Called(ctx, entries)
	return args.Error(0)
}

func (m *MockStorage) QueryLogs(ctx context.Context, filter storage.QueryFilter) ([]storage.LogEntry, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]storage.LogEntry), args.Error(1)
}

func (m *MockStorage) DeleteOldLogs(ctx context.Context, olderThan time.Duration) (int64, error) {
	args := m.Called(ctx, olderThan)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockStorage) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockStorage) HealthCheck(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func TestLogAggregatorServer_SendLogs(t *testing.T) {
	t.Run("successful send", func(t *testing.T) {
		mockStorage := new(MockStorage)
		logger := logrus.New()
		logger.SetOutput(io.Discard) // Suppress log output in tests

		srv := server.NewLogAggregatorServer(mockStorage, logger)

		batch := &pb.LogBatch{
			Events: []*pb.LogEvent{
				{
					Timestamp: time.Now().Format(time.RFC3339Nano),
					Service:   "api",
					Level:     "info",
					Message:   "Test message",
					Fields:    map[string]string{"key": "value"},
				},
			},
		}

		mockStorage.On("InsertLogs", mock.Anything, mock.AnythingOfType("[]storage.LogEntry")).Return(nil)

		resp, err := srv.SendLogs(context.Background(), batch)
		require.NoError(t, err)
		assert.True(t, resp.Ok)
		mockStorage.AssertExpectations(t)
	})

	t.Run("empty batch", func(t *testing.T) {
		mockStorage := new(MockStorage)
		logger := logrus.New()
		logger.SetOutput(io.Discard)

		srv := server.NewLogAggregatorServer(mockStorage, logger)

		batch := &pb.LogBatch{Events: []*pb.LogEvent{}}

		resp, err := srv.SendLogs(context.Background(), batch)
		require.NoError(t, err)
		assert.True(t, resp.Ok)

		// Should not call InsertLogs for empty batch
		mockStorage.AssertNotCalled(t, "InsertLogs")
	})

	t.Run("multiple logs", func(t *testing.T) {
		mockStorage := new(MockStorage)
		logger := logrus.New()
		logger.SetOutput(io.Discard)

		srv := server.NewLogAggregatorServer(mockStorage, logger)

		batch := &pb.LogBatch{
			Events: []*pb.LogEvent{
				{
					Timestamp: time.Now().Format(time.RFC3339Nano),
					Service:   "api",
					Level:     "info",
					Message:   "Message 1",
				},
				{
					Timestamp: time.Now().Format(time.RFC3339Nano),
					Service:   "worker",
					Level:     "debug",
					Message:   "Message 2",
				},
			},
		}

		mockStorage.On("InsertLogs", mock.Anything, mock.MatchedBy(func(entries []storage.LogEntry) bool {
			return len(entries) == 2
		})).Return(nil)

		resp, err := srv.SendLogs(context.Background(), batch)
		require.NoError(t, err)
		assert.True(t, resp.Ok)
		mockStorage.AssertExpectations(t)
	})
}

func TestLogAggregatorServer_QueryLogs(t *testing.T) {
	t.Run("successful query", func(t *testing.T) {
		mockStorage := new(MockStorage)
		logger := logrus.New()
		logger.SetOutput(io.Discard)

		srv := server.NewLogAggregatorServer(mockStorage, logger)

		expectedEntries := []storage.LogEntry{
			{
				ID:        1,
				Timestamp: time.Now(),
				Service:   "api",
				Level:     "info",
				Message:   "Test log",
				Fields:    map[string]string{"key": "value"},
			},
		}

		mockStorage.On("QueryLogs", mock.Anything, mock.MatchedBy(func(filter storage.QueryFilter) bool {
			return filter.Service == "api" && filter.Level == "info"
		})).Return(expectedEntries, nil)

		req := &pb.QueryRequest{
			Service: "api",
			Level:   "info",
		}

		resp, err := srv.QueryLogs(context.Background(), req)
		require.NoError(t, err)
		assert.Len(t, resp.Events, 1)
		assert.Equal(t, "api", resp.Events[0].Service)
		assert.Equal(t, "info", resp.Events[0].Level)
		assert.Equal(t, "Test log", resp.Events[0].Message)
		mockStorage.AssertExpectations(t)
	})

	t.Run("empty results", func(t *testing.T) {
		mockStorage := new(MockStorage)
		logger := logrus.New()
		logger.SetOutput(io.Discard)

		srv := server.NewLogAggregatorServer(mockStorage, logger)

		mockStorage.On("QueryLogs", mock.Anything, mock.Anything).Return([]storage.LogEntry{}, nil)

		req := &pb.QueryRequest{
			Service: "nonexistent",
			Level:   "debug",
		}

		resp, err := srv.QueryLogs(context.Background(), req)
		require.NoError(t, err)
		assert.Empty(t, resp.Events)
		mockStorage.AssertExpectations(t)
	})
}

func TestLogAggregatorServer_Close(t *testing.T) {
	mockStorage := new(MockStorage)
	logger := logrus.New()
	logger.SetOutput(nil)

	srv := server.NewLogAggregatorServer(mockStorage, logger)

	mockStorage.On("Close").Return(nil)

	err := srv.Close()
	assert.NoError(t, err)
	mockStorage.AssertExpectations(t)
}
