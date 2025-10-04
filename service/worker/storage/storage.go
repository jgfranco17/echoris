package storage

import (
	"context"
	"time"
)

// LogEntry represents a log entry in the database
type LogEntry struct {
	ID        int64
	Timestamp time.Time
	Service   string
	Level     string
	Message   string
	Fields    map[string]string
	CreatedAt time.Time
}

// QueryFilter represents filters for querying logs
type QueryFilter struct {
	Service   string
	Level     string
	StartTime *time.Time
	EndTime   *time.Time
	Limit     int
	Offset    int
}

// Storage defines the interface for log storage operations
type Storage interface {
	// InsertLog inserts a single log entry
	InsertLog(ctx context.Context, entry LogEntry) error

	// InsertLogs inserts multiple log entries in a batch
	InsertLogs(ctx context.Context, entries []LogEntry) error

	// QueryLogs retrieves logs based on filters
	QueryLogs(ctx context.Context, filter QueryFilter) ([]LogEntry, error)

	// DeleteOldLogs deletes logs older than the specified duration
	DeleteOldLogs(ctx context.Context, olderThan time.Duration) (int64, error)

	// Close closes the storage connection
	Close() error

	// HealthCheck verifies the storage connection is healthy
	HealthCheck(ctx context.Context) error
}
