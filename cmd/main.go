package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/riyadennis/ingestion-service/server"
	"github.com/riyadennis/ingestion-service/storage"
	"github.com/sirupsen/logrus"
)

func main() {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)
	cf := storage.NewEnvConfig(logger)
	ctx := context.Background()

	client, err := storage.NewClient(cf)
	if err != nil {
		logger.Fatalf("falied to initialise storage client: %v", err)
	}

	err = cf.MakeBucket(ctx, client)
	if err != nil {
		logger.Fatalf("failed to make bucket: %v", err)
	}

	restServer, err := server.NewServer(os.Getenv("PORT"))
	if err != nil {
		logger.Fatalf("failed to initialise server: %v", err)
	}
	logger.Info("Server created")

	signal.Notify(restServer.ShutDown, os.Interrupt, syscall.SIGTERM)

	err = restServer.Run(logger, client, cf.BucketName)
	if err != nil {
		logger.Fatalf("failed to start server: %v", err)
	}

}
