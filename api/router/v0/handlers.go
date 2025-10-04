package v0

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jgfranco17/echoris/api/events"
	"github.com/jgfranco17/echoris/api/httperror"
	"github.com/jgfranco17/echoris/api/logging"
	"github.com/sirupsen/logrus"
)

type HttpHandler func(c *gin.Context) error

func getLogs(client LogClient) HttpHandler {
	return func(c *gin.Context) error {
		logger := logging.FromContext(c)
		service := c.Query("service")
		level := c.Query("level")

		logger.WithFields(logrus.Fields{
			"service": service,
			"level":   level,
		}).Info("Fetching logs")

		logs, err := client.FetchLogs(context.Background(), service, level)
		if err != nil {
			logger.WithError(err).Error("Failed to fetch logs")
			return httperror.New(c, http.StatusInternalServerError, "failed to fetch logs")
		}

		logger.WithField("count", len(logs)).Info("Successfully fetched logs")

		c.JSON(http.StatusOK, gin.H{
			"logs": logs,
		})
		return nil
	}
}

func postLogs(client LogClient) HttpHandler {
	return func(c *gin.Context) error {
		var batch []events.Entry
		if err := c.ShouldBindJSON(&batch); err != nil {
			return httperror.New(c, http.StatusBadRequest, "invalid JSON body")
		}

		// Get logger from context
		logger := logging.FromContext(c)

		logger.WithField("count", len(batch)).Info("Forwarding logs")

		// Forward to gRPC worker service
		if err := client.ForwardLogs(context.Background(), batch); err != nil {
			logger.WithError(err).Error("Failed to forward logs")
			return httperror.New(c, http.StatusInternalServerError, "failed to forward logs")
		}

		logger.Info("Successfully forwarded logs")

		c.JSON(http.StatusOK, gin.H{
			"message": "Logs forwarded successfully",
		})
		return nil
	}
}
