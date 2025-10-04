package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

// PostgresStorage implements Storage interface using PostgreSQL
type PostgresStorage struct {
	db *sql.DB
}

// NewPostgresStorage creates a new PostgreSQL storage instance
func NewPostgresStorage(connString string) (*PostgresStorage, error) {
	db, err := sql.Open("postgres", connString)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetConnMaxIdleTime(1 * time.Minute)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	storage := &PostgresStorage{db: db}

	// Initialize schema
	if err := storage.initSchema(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return storage, nil
}

// initSchema creates the necessary tables
func (s *PostgresStorage) initSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS logs (
		id BIGSERIAL PRIMARY KEY,
		timestamp TIMESTAMPTZ NOT NULL,
		service TEXT NOT NULL,
		level TEXT NOT NULL,
		message TEXT NOT NULL,
		fields JSONB,
		created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_logs_timestamp ON logs(timestamp);
	CREATE INDEX IF NOT EXISTS idx_logs_service ON logs(service);
	CREATE INDEX IF NOT EXISTS idx_logs_level ON logs(level);
	CREATE INDEX IF NOT EXISTS idx_logs_service_level ON logs(service, level);
	CREATE INDEX IF NOT EXISTS idx_logs_created_at ON logs(created_at);
	CREATE INDEX IF NOT EXISTS idx_logs_fields ON logs USING gin(fields);
	`

	_, err := s.db.Exec(schema)
	return err
}

// InsertLog inserts a single log entry
func (s *PostgresStorage) InsertLog(ctx context.Context, entry LogEntry) error {
	fieldsJSON, err := json.Marshal(entry.Fields)
	if err != nil {
		return fmt.Errorf("failed to marshal fields: %w", err)
	}

	query := `
		INSERT INTO logs (timestamp, service, level, message, fields)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err = s.db.ExecContext(ctx, query,
		entry.Timestamp,
		entry.Service,
		entry.Level,
		entry.Message,
		fieldsJSON,
	)

	if err != nil {
		return fmt.Errorf("failed to insert log: %w", err)
	}

	return nil
}

// InsertLogs inserts multiple log entries in a batch
func (s *PostgresStorage) InsertLogs(ctx context.Context, entries []LogEntry) error {
	if len(entries) == 0 {
		return nil
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO logs (timestamp, service, level, message, fields)
		VALUES ($1, $2, $3, $4, $5)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, entry := range entries {
		fieldsJSON, err := json.Marshal(entry.Fields)
		if err != nil {
			return fmt.Errorf("failed to marshal fields: %w", err)
		}

		_, err = stmt.ExecContext(ctx,
			entry.Timestamp,
			entry.Service,
			entry.Level,
			entry.Message,
			fieldsJSON,
		)
		if err != nil {
			return fmt.Errorf("failed to insert log: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// QueryLogs retrieves logs based on filters
func (s *PostgresStorage) QueryLogs(ctx context.Context, filter QueryFilter) ([]LogEntry, error) {
	query := `SELECT id, timestamp, service, level, message, fields, created_at FROM logs WHERE 1=1`
	args := []interface{}{}
	argCount := 1

	if filter.Service != "" {
		query += fmt.Sprintf(" AND service = $%d", argCount)
		args = append(args, filter.Service)
		argCount++
	}

	if filter.Level != "" {
		query += fmt.Sprintf(" AND level = $%d", argCount)
		args = append(args, filter.Level)
		argCount++
	}

	if filter.StartTime != nil {
		query += fmt.Sprintf(" AND timestamp >= $%d", argCount)
		args = append(args, filter.StartTime)
		argCount++
	}

	if filter.EndTime != nil {
		query += fmt.Sprintf(" AND timestamp <= $%d", argCount)
		args = append(args, filter.EndTime)
		argCount++
	}

	query += " ORDER BY timestamp DESC"

	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argCount)
		args = append(args, filter.Limit)
		argCount++
	}

	if filter.Offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argCount)
		args = append(args, filter.Offset)
		argCount++
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query logs: %w", err)
	}
	defer rows.Close()

	var entries []LogEntry
	for rows.Next() {
		var entry LogEntry
		var fieldsJSON []byte

		err := rows.Scan(
			&entry.ID,
			&entry.Timestamp,
			&entry.Service,
			&entry.Level,
			&entry.Message,
			&fieldsJSON,
			&entry.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		if len(fieldsJSON) > 0 {
			if err := json.Unmarshal(fieldsJSON, &entry.Fields); err != nil {
				return nil, fmt.Errorf("failed to unmarshal fields: %w", err)
			}
		}

		entries = append(entries, entry)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return entries, nil
}

// DeleteOldLogs deletes logs older than the specified duration
func (s *PostgresStorage) DeleteOldLogs(ctx context.Context, olderThan time.Duration) (int64, error) {
	cutoffTime := time.Now().Add(-olderThan)

	result, err := s.db.ExecContext(ctx,
		"DELETE FROM logs WHERE created_at < $1",
		cutoffTime,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to delete old logs: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return rowsAffected, nil
}

// Close closes the database connection
func (s *PostgresStorage) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

// HealthCheck verifies the database connection is healthy
func (s *PostgresStorage) HealthCheck(ctx context.Context) error {
	return s.db.PingContext(ctx)
}
