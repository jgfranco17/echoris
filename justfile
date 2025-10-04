# Echoris: Development scripts

INSTALL_PATH := "$HOME/.local"

# Default command
_default:
    @just --list --unsorted

# Sync Go modules
tidy:
    go mod tidy
    @echo "All modules synced, Go workspace ready!"

# CLI local run wrapper
echoris *args:
    @go run ./cli/cmd {{ args }}

# Start the API server
start:
    @go run ./api/cmd --port 8000 --dev

# Run all BDD tests
test:
    @echo "Running unit tests!"
    go clean -testcache
    go test -cover ./...

protos:
    protoc \
        --go_out=. --go_opt=paths=source_relative \
        --go-grpc_out=. --go-grpc_opt=paths=source_relative \
        service/protos/log.proto

# Build the binary
build:
    #!/usr/bin/env bash
    # Detect OS and architecture
    case "$(uname -s)" in
        Linux*) OS="linux" ;;
        Darwin*) OS="darwin" ;;
        *) echo "Error: Unsupported OS (${OS})"; exit 1 ;;
    esac
    case "$(uname -m)" in
        x86_64) ARCH="amd64" ;;
        aarch64) ARCH="arm64" ;;
        arm64) ARCH="arm64" ;;
        *) echo "Error: Unsupported architecture (${ENV_ARCH})"; exit 1 ;;
    esac

    echo "Building devops for ${OS}/${ARCH}..."
    go mod download all
    CGO_ENABLED=0 GOOS="${OS}" GOARCH="${ARCH}" go build -o ./cli/cmd .
    echo "Built binary for devops successfully!"

# Update the project dependencies
update-deps:
    @echo "Updating project dependencies..."
    go get -u ./...
    go mod tidy

# Start the Docker services
up:
    @docker compose up

# Stop the Docker services
down:
    @docker compose down -v
