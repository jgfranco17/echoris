# ========== SERVICE BUILD STAGE ==========
ARG GO_VERSION=1.24
FROM golang:${GO_VERSION}-alpine AS build

RUN apk add --no-cache git make

WORKDIR /src
RUN --mount=type=cache,target=/go/pkg/mod/ \
    --mount=type=bind,source=go.mod,target=go.mod \
    go mod download -x
COPY . .
RUN --mount=type=cache,target=/go/pkg/mod/ \
    CGO_ENABLED=0 go build -a -installsuffix cgo -o worker ./service/worker/cmd

# ========== SERVICE RUNTIME STAGE ==========
FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /app

# Copy binary from builder
COPY --from=build /src/worker .

EXPOSE 50051

ENTRYPOINT ["/app/worker"]
CMD ["-use-env", "-log-level", "info"]
