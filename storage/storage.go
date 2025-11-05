package storage

import (
	"context"
	"os"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/sirupsen/logrus"
)

type Config struct {
	Endpoint   string
	AccessKey  string
	SecretKey  string
	UseSSL     bool
	BucketName string
	Logger     *logrus.Logger
}

func NewEnvConfig(logger *logrus.Logger) Config {
	return Config{
		Endpoint:   os.Getenv("STORAGE_ENDPOINT"),
		AccessKey:  os.Getenv("STORAGE_ACCESS_KEY"),
		SecretKey:  os.Getenv("STORAGE_SECRET_KEY"),
		UseSSL:     os.Getenv("STORAGE_USE_SSL") == "true",
		BucketName: os.Getenv("STORAGE_BUCKET_NAME"),
		Logger:     logger,
	}
}

func NewClient(cfg Config) (*minio.Client, error) {
	// Works with MinIO, GCS, S3, R2, etc.
	return minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})
}

func (cfg *Config) MakeBucket(ctx context.Context, client *minio.Client) error {
	exists, err := client.BucketExists(ctx, cfg.BucketName)
	if err != nil {
		cfg.Logger.Errorf("Error checking if bucket exists: %v", err)
		return err
	}
	if exists {
		cfg.Logger.Infof("Bucket %s already exists", cfg.BucketName)
		return nil
	}
	return client.MakeBucket(ctx, cfg.BucketName, minio.MakeBucketOptions{
		Region:      "us-east-1",
		ForceCreate: true,
	})
}
