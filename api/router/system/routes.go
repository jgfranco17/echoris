package system

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var startTime time.Time

func init() {
	startTime = time.Now()
}

func SetSystemRoutes(route *gin.Engine) {
	startTime = time.Now()
	route.GET("/healthz", HealthCheckHandler())
	route.GET("/metrics", gin.WrapH(promhttp.Handler()))
	route.GET("/service-info", ServiceInfoHandler(startTime))
	for _, homeRoute := range []string{"", "/home"} {
		route.GET(homeRoute, HomeHandler)
	}
	route.NoRoute(NotFoundHandler())
}
