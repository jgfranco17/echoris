package v0

import (
	"github.com/jgfranco17/echoris/api/httperror"

	"github.com/gin-gonic/gin"
)

// Adds v0 routes to the router.
func SetRoutes(route *gin.Engine) error {
	v0 := route.Group("/v0")
	v0.GET("/logs", httperror.WithErrorHandling(getLogs()))
	return nil
}
