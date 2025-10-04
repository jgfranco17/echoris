# Echoris

## Architecture

The system consists of three main components:

- **API Service** (port 8000): HTTP/REST API for log ingestion and querying
- **Worker Service** (port 50051): gRPC service for log storage and retrieval
- **PostgreSQL** (port 5432): Database for persistent log storage

## Quick Start with Docker Compose (Recommended)

The easiest way to get started is using Docker Compose, which sets up all services:

```bash
# From project root
docker-compose up -d

# View logs
docker-compose logs -f api
docker-compose logs -f worker

# Check service status
docker-compose ps

# Test the API
curl http://localhost:8000/healthz

# Send logs via API
curl -X POST http://localhost:8000/v0/logs \
  -H "Content-Type: application/json" \
  -d '[{"timestamp":"2025-01-01T12:00:00Z","service":"test","level":"info","message":"Hello"}]'

# Query logs via API
curl http://localhost:8000/v0/logs?service=test&level=info

# Stop services
docker-compose down

# Stop and remove all data
docker-compose down -v
```

## Local Development Setup

### Prerequisites

1. **PostgreSQL**: Install PostgreSQL 12 or later
2. **Go**: Go 1.24 or later

### Setup PostgreSQL

#### Using Docker (Recommended for Development)

```bash
# Start PostgreSQL in Docker
docker run -d \
  --name echoris-postgres \
  -e POSTGRES_PASSWORD=postgres \
  -e POSTGRES_DB=logs \
  -p 5432:5432 \
  postgres:16-alpine

# Verify it's running
docker ps | grep echoris-postgres
```

#### Using Local PostgreSQL

```bash
# Create database
createdb logs

# Or using psql
psql -U postgres -c "CREATE DATABASE logs;"
```

### Run the Worker Service

```bash
# From the project root
go run ./service/worker/cmd/main.go

# Or with custom connection
go run ./service/worker/cmd/main.go \
  -conn "host=localhost port=5432 user=postgres password=postgres dbname=logs sslmode=disable"

# Using environment variables
export POSTGRES_PASSWORD=postgres
go run ./service/worker/cmd/main.go -use-env
```

## Configuration

### Command-Line Flags

- `-port` - gRPC server port (default: 50051)
- `-conn` - PostgreSQL connection string
- `-log-level` - Log level: debug, info, warn, error (default: info)
- `-use-env` - Use environment variables for configuration

### Environment Variables

Set these when using `-use-env` flag:

```bash
export POSTGRES_HOST=localhost
export POSTGRES_PORT=5432
export POSTGRES_USER=postgres
export POSTGRES_PASSWORD=your_password
export POSTGRES_DB=logs
export POSTGRES_SSLMODE=disable
```

## Testing the Service

### Using grpcurl

```bash
# Install grpcurl
go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest

# List available services
grpcurl -plaintext localhost:50051 list

# Send logs
grpcurl -plaintext -d '{
  "events": [
    {
      "timestamp": "2025-01-01T12:00:00Z",
      "service": "api",
      "level": "info",
      "message": "Test log entry",
      "fields": {"user": "test"}
    }
  ]
}' localhost:50051 logaggregator.LogAggregator/SendLogs

# Query logs
grpcurl -plaintext -d '{
  "service": "api",
  "level": "info"
}' localhost:50051 logaggregator.LogAggregator/QueryLogs
```
