package logging

import (
	"context"

	"github.com/sirupsen/logrus"
)

type contextLogKey string

// List of keys needed by the core functionality
const (
	RequestId   string = "requestId"
	Version     string = "version"
	Environment string = "environment"
	Origin      string = "origin"
	Logger      string = "logger"
)

// FromContext retrieves a logger from the context
func FromContext(ctx context.Context) *logrus.Logger {
	if logger, ok := ctx.Value(Logger).(*logrus.Logger); ok {
		return logger
	}
	panic("no logger set in context")
}
