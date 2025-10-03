package main

import (
	"flag"
	"time"

	env "github.com/jgfranco17/echoris/api/environment"
	"github.com/jgfranco17/echoris/api/router"
	"github.com/jgfranco17/echoris/api/router/system"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

var (
	port    = flag.Int("port", 8080, "Port to listen on")
	devMode = flag.Bool("dev", true, "Run server in debug mode")
)

func init() {
	if env.IsLocalEnvironment() {
		logrus.SetFormatter(&logrus.TextFormatter{
			TimestampFormat:  time.DateTime,
			PadLevelText:     true,
			QuoteEmptyFields: true,
		})
		gin.SetMode(gin.DebugMode)
	} else {
		logrus.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: time.DateTime,
		})
		gin.SetMode(gin.ReleaseMode)
	}
	prometheus.Register(system.HttpLastRequestReceivedTime)
}

func main() {
	flag.Parse()
	var logLevel logrus.Level
	if *devMode {
		logLevel = logrus.DebugLevel
		logrus.Infof("Running API server on port %d in dev mode", *port)
	} else {
		logLevel = logrus.InfoLevel
		gin.SetMode(gin.ReleaseMode)
		logrus.Infof("Running API production server on port %d", *port)
	}

	service, err := router.CreateNewService(*port, logLevel)
	if err != nil {
		logrus.Fatalf("Error creating the server: %v", err)
	}
	err = service.Run()
	if err != nil {
		logrus.Fatalf("Error starting the server: %v", err)
	}
}
