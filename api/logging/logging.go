package logging

import (
	"io"
	"os"
	"strings"
	"time"

	env "github.com/jgfranco17/echoris/api/environment"
	"github.com/sirupsen/logrus"
)

func New(stream io.Writer) *logrus.Logger {
	logger := configureLoggerFromEnv()
	logger.SetOutput(stream)
	return logger
}

func configureLoggerFromEnv() *logrus.Logger {
	logger := logrus.New()

	if os.Getenv(env.ENV_KEY_ENVIRONMENT) == env.APPLICATION_ENV_LOCAL {
		logger.SetFormatter(&logrus.TextFormatter{
			TimestampFormat: time.DateTime,
		})
	} else {
		logger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: time.DateTime,
		})
	}

	var level logrus.Level
	switch strings.ToLower(os.Getenv(env.ENV_KEY_LOG_LEVEL)) {
	case "debug":
		level = logrus.DebugLevel
	case "info":
		level = logrus.InfoLevel
	case "warn":
		level = logrus.WarnLevel
	case "error":
		level = logrus.ErrorLevel
	default:
		level = logrus.InfoLevel
	}
	logger.SetLevel(level)

	return logger
}
