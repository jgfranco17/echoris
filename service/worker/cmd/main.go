package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	pb "github.com/jgfranco17/echoris/service/protos"
	"github.com/jgfranco17/echoris/service/worker/server"
	"github.com/jgfranco17/echoris/service/worker/storage"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

func main() {
	port := flag.Int("port", 50051, "The server port")
	connString := flag.String("conn", "", "PostgreSQL connection string")
	logLevel := flag.String("log-level", "info", "Log level (debug, info, warn, error)")
	useEnv := flag.Bool("use-env", false, "Use environment variables for storage configuration")
	flag.Parse()

	// Configure logger
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	level, err := logrus.ParseLevel(*logLevel)
	if err != nil {
		logger.Warn("Invalid log level, defaulting to info")
		level = logrus.InfoLevel
	}
	logger.SetLevel(level)

	logger.WithFields(logrus.Fields{
		"port": *port,
	}).Info("Starting log aggregator service")

	// Initialize database storage with dependency injection
	var store storage.Storage
	if *useEnv {
		// Load configuration from environment variables
		logger.Info("Loading storage configuration from environment")
		store, err = storage.NewStorageFromEnv()
	} else {
		// Use command-line arguments
		conn := *connString
		if conn == "" {
			logger.Fatal("No PostgreSQL connection defined")
		}

		config := storage.Config{
			ConnString: conn,
		}

		store, err = storage.NewStorage(config)
	}

	if err != nil {
		logger.WithError(err).Fatal("Failed to initialize storage")
	}
	defer store.Close()

	// Verify database connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := store.HealthCheck(ctx); err != nil {
		logger.WithError(err).Fatal("Database health check failed")
	}
	logger.Info("Database connection established")

	// Create gRPC server
	grpcServer := grpc.NewServer()
	logServer := server.NewLogAggregatorServer(store, logger)
	pb.RegisterLogAggregatorServer(grpcServer, logServer)

	// Start listening
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		logger.WithError(err).Fatal("Failed to listen")
	}

	// Handle graceful shutdown
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		logger.WithField("port", *port).Info("Server listening")
		if err := grpcServer.Serve(listener); err != nil {
			logger.WithError(err).Fatal("Failed to serve")
		}
	}()

	// Wait for shutdown signal
	<-done
	logger.Info("Shutting down server...")

	// Graceful shutdown
	grpcServer.GracefulStop()
	logServer.Close()

	logger.Info("Server stopped")
}
