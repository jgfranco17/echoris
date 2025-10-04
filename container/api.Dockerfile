# ========== API BUILD STAGE ==========
ARG GO_VERSION=1.24
FROM golang:${GO_VERSION}-alpine AS build

RUN apk add --no-cache git make

WORKDIR /src
RUN --mount=type=cache,target=/go/pkg/mod/ \
    --mount=type=bind,source=go.mod,target=go.mod \
    go mod download -x
COPY . .
RUN --mount=type=cache,target=/go/pkg/mod/ \
    CGO_ENABLED=0 go build -a -installsuffix cgo -o backend ./api/cmd

# ========== API RUNTIME STAGE ==========
FROM gcr.io/distroless/base-debian12:debug-nonroot

WORKDIR /app
COPY --from=build /src/backend .

EXPOSE 8000
ENV ENVIRONMENT=dev
ENV LOG_LEVEL=debug

ENTRYPOINT ["/app/backend"]
CMD ["--port", "8000"]
