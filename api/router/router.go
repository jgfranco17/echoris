package router

import (
	"fmt"
	"os"

	env "github.com/jgfranco17/echoris/api/environment"
	"github.com/jgfranco17/echoris/api/logging"
	"github.com/jgfranco17/echoris/api/router/headers"
	system "github.com/jgfranco17/echoris/api/router/system"
	v0 "github.com/jgfranco17/echoris/api/router/v0"
	"github.com/sirupsen/logrus"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Service struct {
	Router *gin.Engine
	Port   int
}

func (s *Service) Run() error {
	err := s.Router.Run(fmt.Sprintf(":%v", s.Port))
	if err != nil {
		return fmt.Errorf("Failed to start service on port %v: %w", s.Port, err)
	}
	return nil
}

// Add the fields we want to expose in the logger to the request context
func setupLogger(logger *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(string(logging.Logger), logger)
		if !env.IsLocalEnvironment() {
			requestID := uuid.NewString()
			environment := os.Getenv(env.ENV_KEY_ENVIRONMENT)
			version := os.Getenv(env.ENV_KEY_VERSION)

			// Go recommends contexts to use custom types instead
			// of strings, but Gin defines key as a string.
			c.Set(string(logging.RequestId), requestID)
			c.Set(string(logging.Environment), environment)
			c.Set(string(logging.Version), version)

			originInfo, err := headers.CreateOriginInfoHeader(c)

			if err == nil && originInfo.Origin != "" {
				c.Set(string(logging.Origin), fmt.Sprintf("%s@%s", originInfo.Origin, originInfo.Version))
			}
		}
		c.Next()
	}
}

// Log the start and completion of a request
func logRequest() gin.HandlerFunc {
	return func(c *gin.Context) {
		logger := logging.FromContext(c)

		logger.WithFields(logrus.Fields{
			"method": c.Request.Method,
			"url":    c.Request.URL.Path,
		}).Info("Request started")

		c.Next()

		logger.WithFields(logrus.Fields{
			"status": c.Writer.Status(),
		}).Info("Request completed")
	}
}

// Configure the router adding routes and middlewares
func getRouter() (*gin.Engine, error) {
	logger := logging.New(os.Stdout)
	router := gin.Default()

	router.Use(setupLogger(logger))
	router.Use(logRequest())
	router.Use(system.PrometheusMiddleware())
	system.SetSystemRoutes(router)

	// Create gRPC client
	workerAddr := os.Getenv("WORKER_SERVICE_HOST")
	if workerAddr == "" {
		workerAddr = "localhost"
	}
	workerPort := os.Getenv("WORKER_SERVICE_PORT")
	if workerPort == "" {
		workerPort = "50051"
	}

	grpcAddr := fmt.Sprintf("%s:%s", workerAddr, workerPort)
	logger.WithField("grpc_address", grpcAddr).Info("Connecting to worker service")

	client, err := v0.NewGRPCLogClient(grpcAddr)
	if err != nil {
		logger.WithError(err).Error("Failed to create gRPC client")
		return nil, fmt.Errorf("Failed to create gRPC client: %w", err)
	}

	err = v0.SetRoutes(router, client)
	if err != nil {
		return nil, fmt.Errorf("Failed to set v0 routes: %w", err)
	}

	return router, nil
}

/*
Create a backend service instance.

[IN] port: server port to listen on

[OUT] *Service: new backend service instance
*/
func CreateNewService(port int, logLevel logrus.Level) (*Service, error) {
	router, err := getRouter()
	if err != nil {
		return nil, fmt.Errorf("Failed to create new service instance: %w", err)
	}
	return &Service{
		Router: router,
		Port:   port,
	}, nil
}
