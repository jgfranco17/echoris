package v0

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jgfranco17/echoris/api/events"
	"github.com/jgfranco17/echoris/api/httperror"
)

type HttpHandler func(c *gin.Context) error

func getLogs() HttpHandler {
	return func(c *gin.Context) error {
		c.JSON(http.StatusOK, gin.H{
			"message": "Logs fetched successfully",
		})
		return nil
	}
}

func postLogs() HttpHandler {
	return func(c *gin.Context) error {
		var batch []events.Entry
		if err := c.ShouldBindJSON(&batch); err != nil {
			return httperror.New(c, http.StatusBadRequest, "invalid JSON body")
		}

		// Forward to gRPC worker service
		if err := forwardLogs(batch); err != nil {
			return httperror.New(c, http.StatusInternalServerError, "failed to forward logs")
		}
		return nil
	}
}
