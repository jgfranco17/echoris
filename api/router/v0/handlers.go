package v0

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func getLogs() func(c *gin.Context) error {
	return func(c *gin.Context) error {

		c.JSON(http.StatusOK, gin.H{
			"message": "Logs fetched successfully",
		})
		return nil
	}
}
