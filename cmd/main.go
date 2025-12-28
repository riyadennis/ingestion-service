package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/riyadennis/ingestion-service/business"
	"github.com/riyadennis/ingestion-service/graph"
	"github.com/riyadennis/ingestion-service/server"
	"github.com/riyadennis/ingestion-service/storage"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func main() {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)
	cf := storage.NewEnvConfig(logger)
	ctx := context.Background()

	client, err := storage.NewClient(cf)
	if err != nil {
		logger.Fatalf("failed to initialise storage client: %v", err)
	}

	err = cf.MakeBucket(ctx, client)
	if err != nil {
		logger.Fatalf("failed to make bucket: %v", err)
	}
	bu := business.NewBucketUpload(client, cf.BucketName)
	rootCommand := cobra.Command{Use: "ingestion"}
	restCmd := &cobra.Command{
		Use:   "rest-server",
		Short: "Start REST server",
		Run: func(cmd *cobra.Command, args []string) {
			restServer, err := server.NewServer(os.Getenv("REST_PORT"))
			if err != nil {
				logger.Fatalf("failed to initialise server: %v", err)
			}
			logger.Info("Server created")

			signal.Notify(restServer.ShutDown, os.Interrupt, syscall.SIGTERM)

			err = restServer.Run(logger, bu)
			if err != nil {
				logger.Fatalf("failed to rest start server: %v", err)
			}
		},
	}
	gqlCmd := &cobra.Command{
		Use:   "gql-server",
		Short: "Start graphQL server",
		Run: func(cmd *cobra.Command, args []string) {
			gqlServer := graph.NewServer(
				logger,
				os.Getenv("GQL_PORT"),
				business.NewBucketUpload(client, cf.BucketName),
			)
			signal.Notify(gqlServer.ShutDown, os.Interrupt, syscall.SIGTERM)
			err = gqlServer.Start(os.Getenv("GQL_PORT"))
			if err != nil {
				logger.Fatalf("failed to start graphQL server: %v", err)
			}
			<-gqlServer.ShutDown
		},
	}

	rootCommand.AddCommand(restCmd, gqlCmd)
	err = rootCommand.Execute()
	if err != nil {
		logger.Fatalf("failed to run servers: %v", err)
	}
}
