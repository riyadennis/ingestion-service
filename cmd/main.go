package main

import (
	"context"

	"github.com/riyadennis/ingestion-service/storage"
	"github.com/sirupsen/logrus"
)

func main() {
	logger := logrus.New()
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
}
