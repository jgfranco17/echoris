package v0

import (
	"context"

	"github.com/jgfranco17/echoris/api/events"
	"github.com/stretchr/testify/mock"
)

// MockLogClient is a mock implementation of LogClient for testing
type MockLogClient struct {
	mock.Mock
}

// ForwardLogs sends a batch of log entries (mock implementation)
func (m *MockLogClient) ForwardLogs(ctx context.Context, batch []events.Entry) error {
	args := m.Called(ctx, batch)
	return args.Error(0)
}

// FetchLogs retrieves log entries (mock implementation)
func (m *MockLogClient) FetchLogs(ctx context.Context, service, level string) ([]events.Entry, error) {
	args := m.Called(ctx, service, level)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]events.Entry), args.Error(1)
}

// Close closes the connection (mock implementation)
func (m *MockLogClient) Close() error {
	args := m.Called()
	return args.Error(0)
}
